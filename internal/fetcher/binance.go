package fetcher

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theabgarg/market-aggregator/internal/domain"
)

type BinanceStreamer struct {
	Symbol string
}

type binanceTrade struct {
	Symbol string `json:"s"`
	Price  string `json:"p"`
	Volume string `json:"q"`
}

func (b *BinanceStreamer) Stream(ctx context.Context, symbol string, out chan<- domain.MarketData) error {
	url := fmt.Sprintf("wss://stream.binance.com:9443/ws/%s@trade", b.Symbol)
	log.Printf("[Binance] connecting to url %s...", url)
	conn, _ , err:= websocket.DefaultDialer.DialContext(ctx, url, nil)
	if err != nil {
		log.Printf("[Binance] failed while connecting to url %s...", url)
		return fmt.Errorf("[Binance] connection error: %v", err)
	}

	defer conn.Close()

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	for {
		var trade binanceTrade
		fmt.Println("reading json", trade)
		err := conn.ReadJSON(&trade)
		fmt.Println("reading json 2", trade)
		if err != nil {
			fmt.Println("got error reading a json")
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("[Binance] read error: %w", err)
		}
		price, _ := strconv.ParseFloat(trade.Price, 64)
		volume, _ := strconv.ParseFloat(trade.Volume, 64)

		fmt.Println("outputting data", trade.Symbol)

		out <- domain.MarketData{
			Symbol: trade.Symbol,
			Price:  price,
			Volume: volume,
			Timestamp: time.Now(),
			Source: "Binance",
		}
	}
}
