package domain

import (
	"sync"
	"testing"
)

func TestMarketCache_concurrency(t *testing.T) {
	redisAddr := "localhost:6379"
	cache := NewMarketCache(redisAddr)
	symbol := "BTCUSD"

	cache.Update(symbol, 100.0)

	routines := 100
	var wg sync.WaitGroup
	wg.Add(routines * 2)

	for i := 0; i < routines; i++ {
		go func(val float64) {
			defer wg.Done()
			cache.Update(symbol, val)
		}(float64(i))

		go func() {
			defer wg.Done()
			_, exists := cache.Get(symbol)
			if !exists {
				t.Errorf("Price not found for symbol %s", symbol)
			}
		}()
	}
	wg.Wait()

	if _, exists := cache.Get(symbol); !exists {
		t.Fatalf("Cache was empty after concurrent operations")
	}
}
