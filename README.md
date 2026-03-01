# 📈 Concurrent Market Data Aggregator

A high-performance, real-time backend service built in Go. This application concurrently ingests, normalizes, and serves live market data streams from multiple cryptocurrency exchanges using WebSockets and an idiomatic Fan-In concurrency architecture.

## 🎯 Project Vision
This service acts as the foundational data ingestion layer for a larger trading application ecosystem. By handling the heavy lifting of maintaining persistent WebSocket connections, standardizing disparate exchange data formats, and safely caching the state in memory, it provides a unified, blazing-fast API for downstream consumers (like trading bots, algorithmic engines, or live UI dashboards).

## 🏗️ System Architecture
The core engine relies heavily on Go's concurrency primitives to achieve millisecond-latency data processing without thread-blocking.



### Core Components
* **The Producers (Data Streamers):** Independent, long-lived goroutines that maintain persistent WebSocket connections to external exchanges (e.g., Binance). They normalize incoming JSON payloads into a standard `domain.MarketData` struct.
* **The Hub (Fan-In Pattern):** A buffered Go channel (`chan domain.MarketData`) that acts as a multiplexer, funneling asynchronous price ticks from all producers into a single, ordered stream.
* **The State (Thread-Safe Cache):** A centralized, in-memory cache protected by a `sync.RWMutex`. It allows high-frequency, lock-free reads from HTTP clients while safely serializing the writes coming from the live feed.
* **The Consumer (HTTP API):** A standard library HTTP server that exposes the latest cached market state to external clients, protected by graceful shutdown mechanics.

## ✨ Current Features
- **Real-Time Ingestion:** Native integration with Binance's live trade WebSocket API (`gorilla/websocket`).
- **Concurrent Safety:** Strict data race prevention using Read/Write Mutexes.
- **Context-Aware Lifecycles:** Global use of `context.Context` to propagate cancellation signals and prevent goroutine leaks.
- **Graceful Shutdowns:** OS-level signal handling to gracefully drain active HTTP requests and close WebSocket connections upon termination (`SIGINT`/`SIGTERM`).
- **Environment Configuration:** 12-factor app compliance using `.env` files for dynamic routing and port binding.

## 🛠️ Tech Stack
- **Language:** Go (Golang) 1.22+
- **Routing:** `net/http` (Standard Library)
- **Concurrency:** Goroutines, Channels, `sync.WaitGroup`, `sync.RWMutex`
- **WebSockets:** `github.com/gorilla/websocket`
- **Config:** `github.com/joho/godotenv`

## 🚀 Getting Started

### Prerequisites
* Go installed on your local machine.

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/theabgarg/market-aggregator.git
   cd market-aggregator