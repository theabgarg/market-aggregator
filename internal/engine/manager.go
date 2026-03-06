package engine

import (
	"context"
	"log/slog"
	"sync"

	"github.com/theabgarg/market-aggregator/internal/domain"
	"github.com/theabgarg/market-aggregator/internal/fetcher"
)

type StreamManager struct {
	mu        sync.Mutex
	streams   map[string]context.CancelFunc
	parentCtx context.Context
	liveFeed  chan<- domain.MarketData
}

func NewStreamManager(ctx context.Context, feed chan<- domain.MarketData) *StreamManager {
	return &StreamManager{
		streams:   make(map[string]context.CancelFunc),
		parentCtx: ctx,
		liveFeed:  feed,
	}
}

func (m *StreamManager) Start(symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.streams[symbol]; exists {
		return
	}

	slog.Info("spinning up new binance connection", "symbol", symbol)

	ctx, cancel := context.WithCancel(m.parentCtx)
	m.streams[symbol] = cancel

	streamer := &fetcher.BinanceStreamer{Symbol: symbol}

	go func() {
		err := streamer.Stream(ctx, symbol, m.liveFeed)

		if err != nil && err != context.Canceled {
			slog.Error("dynamic stream crashed", "symbol", symbol, "error", err.Error())
		}

		m.mu.Lock()
		delete(m.streams, symbol)
		m.mu.Unlock()
	}()

}

func (m *StreamManager) Stop(symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cancel, exists := m.streams[symbol]; exists {
		slog.Info("Stopping binance connection", "symbol", symbol)
		cancel()
		delete(m.streams, symbol)
	}
}
