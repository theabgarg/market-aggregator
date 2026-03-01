package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theabgarg/market-aggregator/internal/config"
	"github.com/theabgarg/market-aggregator/internal/domain"
	"github.com/theabgarg/market-aggregator/internal/fetcher"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Responser handles HTTP requests, utilizing dependency injection for the cache.
type Responser struct {
	cache *domain.MarketCache
}

func (a *Responser) handleAggregator(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	price, exist := a.cache.Get(symbol)
	if !exist {
		http.Error(w, "Price not available", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"symbol": symbol,
		"price":  price,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func startEngine(ctx context.Context, cache *domain.MarketCache, broadcaster *domain.Broadcaster, symbol string) {
	streamers := []domain.DataStreamer{
		&fetcher.BinanceStreamer{Symbol: symbol},
	}

	liveFeed := make(chan domain.MarketData, 100)
	var wg sync.WaitGroup

	for _, s := range streamers {
		wg.Add(1)
		go func(streamer domain.DataStreamer) {
			defer wg.Done()

			err := streamer.Stream(ctx, symbol, liveFeed)
			if err != nil && err != context.Canceled {
				slog.Error("stream crashed", "streamer type", fmt.Sprintf("%T", streamer), "symbol", symbol, "error", err.Error())
			}
		}(s)
	}

	go func() {
		wg.Wait()
		slog.Info("All streams have shut down. Closing hub.")
		close(liveFeed)
	}()

	slog.Info("Listening for live trades... (Press Ctrl+C to stop)")
	for tick := range liveFeed {
		cache.Update(tick.Symbol, tick.Price)
		broadcaster.Broadcast(tick)
	}
}

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("starting market data engine", "port", cfg.Port, "target_symbol", cfg.TargetSymbol)

	cache := domain.NewMarketCache()
	broadcaster := domain.NewBroadcaster()
	responser := &Responser{cache: cache}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: http.DefaultServeMux,
	}

	http.HandleFunc("/api/prices", responser.handleAggregator)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Websocket upgrade failed: %v\n", err)
			return
		}

		broadcaster.AddClient(conn)
		defer broadcaster.RemoveClient(conn)

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go startEngine(ctx, cache, broadcaster, cfg.TargetSymbol)

	go func() {
		log.Printf("HTTP API listening on port %s...\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server crashed: %v", err)
		}
	}()

	// 4. Block the Main Thread
	// The program will sit precisely on this line forever, until you press Ctrl+C
	<-ctx.Done()
	slog.Warn("\nShutdown signal received! Initiating graceful shutdown...")

	// 5. Graceful HTTP Shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP Server forced to shutdown due to error: %v", err)
	}

	fmt.Println("Engine successfully shut down. Goodbye!")
}
