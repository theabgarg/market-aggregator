package domain

import "sync"

type MarketCache struct {
	mu     sync.RWMutex
	prices map[string]float64
}

func NewMarketCache() *MarketCache {
	return &MarketCache{
		prices: make(map[string]float64),
	}
}

func (c *MarketCache) Update(symbol string, price float64) {
	c.mu.Lock()         
	defer c.mu.Unlock() 
	c.prices[symbol] = price
}

func (c *MarketCache) Get(symbol string) (float64, bool) {
	c.mu.RLock()         
	defer c.mu.RUnlock() 

	price, exists := c.prices[symbol]
	return price, exists
}