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
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"github.com/theabgarg/market-aggregator/internal/api"
	"github.com/theabgarg/market-aggregator/internal/config"
	"github.com/theabgarg/market-aggregator/internal/domain"
	"github.com/theabgarg/market-aggregator/internal/engine"
	"github.com/theabgarg/market-aggregator/internal/repository"
	"github.com/theabgarg/market-aggregator/internal/telemetry"
)

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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	cache := domain.NewMarketCache(rdb)
	broadcaster := domain.NewBroadcaster(
		func(symbol string) { StreamManager.Start(symbol) },
		func(symbol string) { StreamManager.Stop(symbol) },
	)

	apiHandler := api.NewHandler(cache, repo, broadcaster)

	mux := http.NewServeMux()

	apiHandler.RegisterRoutes(mux)

	mux.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	limiter := api.NewIPRateLimiter(2, 5)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
	})

	handlerChain := c.Handler(limiter.Middleware(api.RequestLogger(mux)))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: handlerChain,
	}

	dbWriteChan := make(chan domain.MarketData, 5000)

	go broadcaster.StartRedisListener(ctx, rdb)

	go func() {
		slog.Info("engine router running...")
		for tick := range liveFeed {
			telemetry.TicksProcessed.Inc()
			cache.Update(tick.Symbol, tick.Price)
			dbWriteChan <- tick
			tickJSON, err := json.Marshal(tick)
			if err != nil {
				slog.Error("failed to marshal tick for redis", "error", err)
				continue
			}

			err = rdb.Publish(context.Background(), "market:"+tick.Symbol, tickJSON).Err()
			if err != nil {
				slog.Error("failed to publish to redis", "symbol", tick.Symbol, "error", err)
			}
		}
	}()

	go func() {
		batchSize := 100
		batch := make([]domain.MarketData, 0, batchSize)

		flushTicker := time.NewTicker(500 * time.Millisecond)
		defer flushTicker.Stop()
		for {
			select {
			case tick, ok := <-dbWriteChan:
				if !ok {
					if len(batch) > 0 {
						repo.InsertTickBatch(context.Background(), batch)
					}
					slog.Info("database worker did final flush and returned")
					return
				}
				batch = append(batch, tick)
				if len(batch) >= batchSize {
					err := repo.InsertTickBatch(context.Background(), batch)
					if err != nil {
						slog.Error("failed batch insert", "error", err.Error())
					}
					batch = batch[:0]
				}
			case <-flushTicker.C:
				if len(batch) > 0 {
					err := repo.InsertTickBatch(context.Background(), batch)
					if err != nil {
						slog.Error("failed timed batch insert", "error", err.Error())
					}
					batch = batch[:0]
				}

			}
		}
	}()

	// go startEngine(ctx, cache, broadcaster, cfg.TargetSymbol, dbWriteChan)

	go func() {
		log.Printf("HTTP API listening on port %s...\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server crashed: %v", err)
		}
	}()

	<-ctx.Done()
	slog.Warn("\nShutdown signal received! Initiating graceful shutdown...")

	broadcaster.Shutdown()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP Server forced to shutdown due to error: %v", err)
	}

	close(dbWriteChan)
	fmt.Println("Engine successfully shut down. Goodbye!")
}
