package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveClients = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "aggregator_active_websocket_clients",
		Help: "the total number of currenctly active websocket clients",
	})

	TicksProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "aggregator_ticks_processed_total",
		Help: "The total number of market ticks ingested from Binance",
	})
)
