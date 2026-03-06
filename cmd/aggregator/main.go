package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/theabgarg/market-aggregator/internal/api"
	"github.com/theabgarg/market-aggregator/internal/config"
	"github.com/theabgarg/market-aggregator/internal/domain"
	"github.com/theabgarg/market-aggregator/internal/engine"
	"github.com/theabgarg/market-aggregator/internal/fetcher"
	"github.com/theabgarg/market-aggregator/internal/repository"
)

func startEngine(ctx context.Context, cache *domain.MarketCache, broadcaster *domain.Broadcaster, symbol string, dbWriteChan chan<- domain.MarketData) {
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
		dbWriteChan <- tick
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

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		slog.Error("database url fetch failed")
		os.Exit(1)
	}

	slog.Info("starting market data engine", "port", cfg.Port, "target_symbol", cfg.TargetSymbol)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	repo, err := repository.NewPostgresRepo(ctx, dbUrl)

	if err != nil {
		slog.Error("Database Intialization Failed", "error", err.Error())
		os.Exit(1)
	}

	defer repo.Close()

	liveFeed := make(chan domain.MarketData, 100)

	StreamManager := engine.NewStreamManager(ctx, liveFeed)

	cache := domain.NewMarketCache()
	broadcaster := domain.NewBroadcaster(
		func(symbol string) { StreamManager.Start(symbol) },
		func(symbol string) { StreamManager.Stop(symbol) },
	)

	apiHandler := api.NewHandler(cache, repo, broadcaster)

	mux := http.NewServeMux()

	apiHandler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mux,
	}

	dbWriteChan := make(chan domain.MarketData, 500)

	go func() {
		slog.Info("engine router running...")
		for tick := range liveFeed {
			fmt.Print(tick, "tick")
			cache.Update(tick.Symbol, tick.Price)
			broadcaster.Broadcast(tick)
			dbWriteChan <- tick
		}
	}()

	go func() {
		for tick := range dbWriteChan {
			insertCtx, cancelInsert := context.WithTimeout(context.Background(), 2*time.Second)
			err := repo.InsertTick(insertCtx, tick)

			if err != nil {
				slog.Error("failed to write tick to DB", "error", err.Error())
			}
			cancelInsert()
		}
	}()

	// go startEngine(ctx, cache, broadcaster, cfg.TargetSymbol, dbWriteChan)

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

	broadcaster.Shutdown()

	// 5. Graceful HTTP Shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP Server forced to shutdown due to error: %v", err)
	}

	close(dbWriteChan)
	fmt.Println("Engine successfully shut down. Goodbye!")
}
