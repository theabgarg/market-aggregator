package domain

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Broadcaster struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (b *Broadcaster) AddClient(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[conn] = true
	conn.WriteMessage(websocket.TextMessage, []byte("welcome to our websocket"))
	log.Printf("New Client connected. total connections: %d\n", len(b.clients))
}

func (b *Broadcaster) RemoveClient(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _,ok := b.clients[conn]; ok {
		delete(b.clients, conn)
		conn.Close()
		log.Printf("Client Disconnected. total active connections: %d\n",  len(b.clients))
	}
}

func (b *Broadcaster) Broadcast(tick MarketData) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for conn := range b.clients {
		err := conn.WriteJSON(tick)
		if err != nil {
			log.Printf("Failed to send to client. removing connection: %v\n", err)
			delete(b.clients, conn)
			conn.Close()
		}
	}
}