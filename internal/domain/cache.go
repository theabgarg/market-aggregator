package domain

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type MarketCache struct {
	client *redis.Client
}

func NewMarketCache(redisAddr string) *MarketCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to redis", "error", err.Error())
		panic(err)
	}
	slog.Info("connected to redis")
	return &MarketCache{client: rdb}
}

func (c *MarketCache) Update(symbol string, price float64) {
	err := c.client.Set(context.Background(), symbol, price, 0).Err()
	if err != nil {
		slog.Error("failed to update redis", "symbol", symbol, "error", err.Error())
	}
}

func (c *MarketCache) Get(symbol string) (float64, bool) {
	price, err := c.client.Get(context.Background(), symbol).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, false
		}
		slog.Error("failed to read from redis cache", "symbol", symbol, "error", err.Error())
		return 0, false
	}

	return price, true
}
