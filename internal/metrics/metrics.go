package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Ingestion Metrics
var (
	EventsIngestedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "relay",
			Subsystem: "ingest",
			Name:      "events_total",
			Help:      "Total number of events received by the ingestion API.",
		},
		[]string{"status", "event_type"},
	)

	// Delivery Worker Metrics
	DeliveriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "relay",
			Subsystem: "worker",
			Name:      "deliveries_total",
			Help:      "Total number of webhook deliveries attempted by workers.",
		},
		[]string{"endpoint_id", "status", "status_code"},
	)

	DeliveryDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "relay",
			Subsystem: "worker",
			Name:      "delivery_duration_seconds",
			Help:      "Latency distribution of HTTP delivery requests to customer destinations.",
			Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"endpoint_id"},
	)

	// Queue Metrics
	QueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "relay",
			Subsystem: "queue",
			Name:      "depth",
			Help:      "Current number of pending delivery tasks in Redis.",
		},
	)
)
