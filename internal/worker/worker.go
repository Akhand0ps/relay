package worker

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/akhand0ps/relay/internal/id"
	"github.com/akhand0ps/relay/internal/limiter"
	"github.com/akhand0ps/relay/internal/metrics"
	"github.com/akhand0ps/relay/internal/models"
	"github.com/akhand0ps/relay/internal/queue"
	"github.com/akhand0ps/relay/internal/signature"
	"github.com/akhand0ps/relay/internal/store"
)

type WorkerPool struct {
	Queue   queue.Queue
	client  *http.Client
	store   store.Store
	limiter *limiter.EndpointLimiter
}

func NewWorkerPool(q queue.Queue, store store.Store, lim *limiter.EndpointLimiter) *WorkerPool {
	return &WorkerPool{
		Queue: q,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		store:   store,
		limiter: lim,
	}
}

// Spawns N worker goroutines in the background continuously consuming from the distributed queue
func (wp *WorkerPool) Start(workerCount int) {
	for i := 1; i <= workerCount; i++ {
		go func(workerId int) {
			for {
				task, err := wp.Queue.Dequeue(context.Background())
				if err != nil {
					// Avoid busy spin on network error or shutdown
					time.Sleep(500 * time.Millisecond)
					continue
				}
				if task != nil {
					// Update queue depth gauge
					if depth, err := wp.Queue.Len(context.Background()); err == nil {
						metrics.QueueDepth.Set(float64(depth))
					}
					wp.processTask(workerId, *task)
				}
			}
		}(i)
	}
}

func (wp *WorkerPool) processTask(workerId int, task models.DeliveryTask) {
	// Apply Token Bucket rate limiting: wait for an available token before firing HTTP request
	if err := wp.limiter.Wait(context.Background(), task.Endpoint.ID); err != nil {
		fmt.Printf("[Worker %d] Rate limit aborted for %s: %v\n", workerId, task.Endpoint.URL, err)
		return
	}

	fmt.Printf("[Worker %d] Delivering %s to %s...\n", workerId, task.Event.Type, task.Endpoint.URL)

	now := time.Now().Unix()
	sig := signature.Sign(task.Endpoint.Secret, now, task.Event.Payload)

	req, err := http.NewRequest("POST", task.Endpoint.URL, bytes.NewReader(task.Event.Payload))
	if err != nil {
		fmt.Printf("[Worker %d] Failed to create request: %v\n", workerId, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Relay-Event-Id", task.Event.ID)
	req.Header.Set("X-Relay-TimeStamp", fmt.Sprintf("%d", now))
	req.Header.Set("X-Relay-Signature", sig)

	// Measure HTTP delivery latency for Prometheus histogram
	startCall := time.Now()
	resp, err := wp.client.Do(req)
	duration := time.Since(startCall).Seconds()
	metrics.DeliveryDurationSeconds.WithLabelValues(task.Endpoint.ID).Observe(duration)

	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
		defer resp.Body.Close()
	}

	var errMessage string
	if err != nil {
		errMessage = err.Error()
	} else {
		errMessage = ""
	}

	attempt := models.DeliveryAttempt{
		ID:           id.GenerateEventId("att"),
		DeliveryId:   task.Delivery.ID,
		StatusCode:   statusCode,
		ErrorMessage: errMessage,
		AttemptedAt:  time.Now(),
	}

	if err := wp.store.RecordDeliveryAttempt(attempt); err != nil {
		fmt.Printf("[Worker %d] Failed to record delivery attempt: %v\n", workerId, err)
	}

	statusStr := fmt.Sprintf("%d", statusCode)

	// Determine if delivery succeeded
	if err == nil && statusCode >= 200 && statusCode < 300 {
		metrics.DeliveriesTotal.WithLabelValues(task.Endpoint.ID, "success", statusStr).Inc()
		fmt.Printf("[Worker %d] ✔ Delivery SUCCESS to %s (Status: %d)\n", workerId, task.Endpoint.URL, statusCode)
	} else {
		try := ShouldRetry(statusCode, err)
		if !try {
			task.Delivery.Status = models.DeliveryStatusFailed
			metrics.DeliveriesTotal.WithLabelValues(task.Endpoint.ID, "failed", statusStr).Inc()
			fmt.Printf("[Worker %d] ❌ Permanent failure, not retrying\n", workerId)
			return
		}

		task.Delivery.AttemptCount++
		if task.Delivery.AttemptCount >= 5 {
			task.Delivery.Status = models.DeliveryStatusDLQ
			metrics.DeliveriesTotal.WithLabelValues(task.Endpoint.ID, "dlq", statusStr).Inc()
			fmt.Printf("[Worker %d] Delivery %s exceeded max retries, sent to DLQ\n", workerId, task.Delivery.ID)
			return
		}

		metrics.DeliveriesTotal.WithLabelValues(task.Endpoint.ID, "retrying", statusStr).Inc()
		delay := CalculateBackoff(task.Delivery.AttemptCount)

		// Wait for the backoff delay in background, then re-enqueue to Redis
		go func(t models.DeliveryTask, d time.Duration) {
			fmt.Printf("[Worker %d] Scheduling retry #%d for %s in %v...\n", workerId, t.Delivery.AttemptCount, t.Endpoint.URL, d)
			time.Sleep(d)
			if err := wp.Queue.Enqueue(context.Background(), t); err != nil {
				fmt.Printf("[Worker %d] Failed to re-enqueue retry to Redis: %v\n", workerId, err)
			}
		}(task, delay)
	}
}
