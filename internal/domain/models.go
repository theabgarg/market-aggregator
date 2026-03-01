package domain

import (
	"context"
	"time"
)

type MarketData struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

type DataFetcher interface {
	Fetch(ctx context.Context,symbol string) (MarketData, error)
}

type DataStreamer interface {
	Stream(ctx context.Context, symbol string, out chan<- MarketData) error
}

