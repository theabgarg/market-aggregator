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

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
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
	// mux.Handle("/api/history", JWTMiddleware(http.HandlerFunc(h.handleHistory)))
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
		respondWithError(w, http.StatusNotFound, "Price not available")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"symbol": symbol,
		"price":  price,
	})
}

func (h *Handler) handleHistory(w http.ResponseWriter, r *http.Request) {
	// userID := r.Context().Value(userIDKey).(string)
	// slog.Info("authorized request", "user ID", userID)
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
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
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
	defer h.broadcaster.RemoveClient(conn)

	for {
		var msg domain.ClientMessage

		err := conn.ReadJSON(&msg)

		if err != nil {
			slog.Info("client disconnected", "reason", err.Error())
			break
		}

		switch msg.Action {
		case "subscribe":
			h.broadcaster.Subscribe(conn, msg.Symbol)
		case "unsubscribe":
			h.broadcaster.Unsubscribe(conn, msg.Symbol)
		default:
			slog.Warn("unknown websocket action received", "action", msg.Action)
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)

	errorObj := ErrorResponse{
		Code:  code,
		Error: message,
	}

	json.NewEncoder(w).Encode(errorObj)
}
