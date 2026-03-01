package fetcher

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/theabgarg/market-aggregator/internal/domain"
)

type MockExchange struct {
	Name string
}

func (m *MockExchange) Fetch(ctx context.Context, symbol string) (domain.MarketData, error) {
	latency := time.Duration(rand.IntN(500)) * time.Millisecond

	timer := time.NewTimer(latency)
	defer timer.Stop()

	select {
	case <-timer.C:
		return domain.MarketData{
			Symbol:    symbol,
			Price:     100.0 + rand.Float64()*10,
			Volume:    rand.Float64() * 1000,
			Timestamp: time.Now(),
			Source:    m.Name,
		}, nil

	case <-ctx.Done():
		return domain.MarketData{}, fmt.Errorf("%s Fetch aborted: %w", m.Name, ctx.Err())
	}
}

type MockWebsocketExchange struct {
	Name string
}

func (m *MockWebsocketExchange) Stream(ctx context.Context, symbol string, out chan<- domain.MarketData) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	currentPrice := 100.0

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] Connection closed\n", m.Name)
			return ctx.Err()
		case <-ticker.C:
			currentPrice += (rand.Float64() * 2) - 1.0
			out <- domain.MarketData{
				Symbol:    symbol,
				Price:     currentPrice + rand.Float64()*10 - 5,
				Volume:    rand.Float64() * 1000,
				Timestamp: time.Now(),
				Source:    m.Name,
			}
		}
	}
}
