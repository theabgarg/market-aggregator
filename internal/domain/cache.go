package domain

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type MarketCache struct {
	client *redis.Client
}

func NewMarketCache(client *redis.Client) *MarketCache {
	return &MarketCache{client: client}
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
