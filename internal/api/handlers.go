package api

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/theabgarg/market-aggregator/internal/domain"
	"github.com/theabgarg/market-aggregator/internal/repository"
)

type Handler struct {
	cache       *domain.MarketCache
	repo        *repository.PostgresRepo
	broadcaster *domain.Broadcaster
	upgrader    websocket.Upgrader
}

func NewHandler(c *domain.MarketCache, r *repository.PostgresRepo, b *domain.Broadcaster) *Handler {
	return &Handler{
		cache:       c,
		repo:        r,
		broadcaster: b,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/prices", h.handleAggregator)
	mux.HandleFunc("/api/history", h.handleHistory)
	mux.HandleFunc("/ws", h.handleWebSocket)
}

func (h *Handler) handleAggregator(w http.ResponseWriter, r *http.Request) {

	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	price, exist := h.cache.Get(symbol)
	if !exist {
		http.Error(w, "Price not available", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"price":  price,
	})
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol is a required field", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	ticks, err := h.repo.GetHistoricalTicks(ctx, symbol, 100)
	if err != nil {
		slog.Error("failed to fetch historical data", "error", err.Error())
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(ticks)

}

func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade failed: %v\n", err)
		return
	}
	h.broadcaster.AddClient(conn)
	defer h.broadcaster.RemoveClient(conn)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
