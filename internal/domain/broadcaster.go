package domain

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	Action string `json:"action"`
	Symbol string `json:"symbol"`
}

type Broadcaster struct {
	mu     sync.Mutex
	topics map[string]map[*websocket.Conn]bool
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		topics: make(map[string]map[*websocket.Conn]bool),
	}
}

func (b *Broadcaster) Subscribe(conn *websocket.Conn, symbol string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.topics[symbol] == nil {
		b.topics[symbol] = make(map[*websocket.Conn]bool)
	}

	b.topics[symbol][conn] = true
	slog.Info("client subscribed", "symbol", symbol, "total_subscribers", len(b.topics[symbol]))
}

func (b *Broadcaster) Unsubscribe(conn *websocket.Conn, symbol string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if clients, exists := b.topics[symbol]; exists {
		delete(clients, conn)
		slog.Info("client unsubscribed", "symbol", symbol, "remaining subs", len(b.topics[symbol]))
	}
}

func (b *Broadcaster) RemoveClient(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for symbol, clients := range b.topics {
		if _, exists := clients[conn]; exists {
			delete(clients, conn)
			slog.Info("client cleaned up from topic", "symbol", symbol)
		}
	}
	conn.Close()
}

func (b *Broadcaster) Broadcast(tick MarketData) {
	b.mu.Lock()
	defer b.mu.Unlock()

	clients, exists := b.topics[tick.Symbol]

	if !exists || len(clients) == 0 {
		return
	}

	for conn := range clients {
		err := conn.WriteJSON(tick)
		if err != nil {
			slog.Error("failed to write to client", "error", err.Error())
		}
	}
}

func (b *Broadcaster) Shutdown() {
	b.mu.Lock()
	defer b.mu.Unlock()

	slog.Info("broadcaster initiating graceful disconnect for all clients")

	closeMessage := websocket.FormatCloseMessage(
		websocket.CloseNormalClosure,
		"Server is shutting down for maintenance",
	)

	for symbol, clients := range b.topics{
		for conn := range clients {
			err := conn.WriteMessage(websocket.CloseMessage, closeMessage)
			if err != nil {
				slog.Warn("Failed to send close message", "symbol", symbol, "error", err.Error())
			}
			conn.Close()
		}
		delete(b.topics, symbol)
	}

	slog.Info("All websocket clients diconnected successfully")
}
