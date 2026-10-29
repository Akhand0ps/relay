package main

import (
	"fmt"
	"net/http"

	"github.com/akhand0ps/relay/internal/api"
	"github.com/akhand0ps/relay/internal/limiter"
	"github.com/akhand0ps/relay/internal/queue"
	"github.com/akhand0ps/relay/internal/store"
	"github.com/akhand0ps/relay/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Initialize PostgreSQL persistent store
	connStr := "postgres://postgres:pgpassword@localhost:5432/relay?sslmode=disable"
	sharedStore, err := store.NewPostgresStore(connStr)
	if err != nil {
		fmt.Printf("Database connection failed: %v\n", err)
		return
	}

	// 2. Initialize Redis distributed queue
	redisAddr := "localhost:6379"
	queueKey := "relay:queue:deliveries"
	redisQueue, err := queue.NewRedisQueue(redisAddr, queueKey)
	if err != nil {
		fmt.Printf("Redis connection failed: %v\n", err)
		return
	}
	defer redisQueue.Close()

	// 3. Initialize Token Bucket per-endpoint rate limiter (10 requests/sec, burst capacity of 20)
	endpointLimiter := limiter.NewEndpointLimiter(10.0, 20)

	// 4. Initialize WorkerPool with RedisQueue, PostgreSQL store, and Token Bucket limiter
	wp := worker.NewWorkerPool(redisQueue, sharedStore, endpointLimiter)
	wp.Start(3) // 3 concurrent worker goroutines

	// 5. Initialize HTTP API handlers
	h := api.NewHandler(sharedStore, wp)

	mux := http.NewServeMux()

	// Observability & Health endpoints
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", api.HealthCheckHandler)
	mux.HandleFunc("/health/live", h.LivenessHandler)
	mux.HandleFunc("/health/ready", h.ReadinessHandler)

	// Webhook API routes
	mux.HandleFunc("POST /v1/events", h.IngestEventHandler)
	mux.HandleFunc("GET /v1/events/{id}", h.GetEventHandler)
	mux.HandleFunc("POST /v1/endpoints", h.CreateEndpointHandler)
	mux.HandleFunc("GET /v1/endpoints", h.ListEndpointHandler)
	mux.HandleFunc("POST /v1/events/replay", h.ReplayHandler)

	port := ":3000"
	fmt.Printf("Relay Server running on %s\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
