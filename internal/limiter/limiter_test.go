package limiter

import (
	"context"
	"testing"
	"time"
)

func TestEndpointLimiterBurstAndRefill(t *testing.T) {
	// 5 tokens/sec, max burst of 3
	el := NewEndpointLimiter(5.0, 3)
	endpointID := "ep_test_shopify"

	// 1. First 3 requests must be allowed immediately (burst capacity)
	for i := 1; i <= 3; i++ {
		if !el.Allow(endpointID) {
			t.Fatalf("expected request %d to be allowed under burst capacity", i)
		}
	}

	// 2. 4th request must be rejected (bucket is empty)
	if el.Allow(endpointID) {
		t.Fatal("expected 4th request to be denied as bucket is empty")
	}

	// 3. Wait for token refill (250ms is enough for 1 token at 5 tokens/sec)
	time.Sleep(250 * time.Millisecond)

	if !el.Allow(endpointID) {
		t.Fatal("expected request to be allowed after token refill")
	}
}

func TestEndpointLimiterIsolation(t *testing.T) {
	// 2 tokens/sec, burst of 1
	el := NewEndpointLimiter(2.0, 1)

	epA := "ep_customer_a"
	epB := "ep_customer_b"

	// Exhaust epA
	if !el.Allow(epA) {
		t.Fatal("expected epA first request to pass")
	}
	if el.Allow(epA) {
		t.Fatal("expected epA second request to be denied")
	}

	// epB must still have its full token available (endpoint isolation)
	if !el.Allow(epB) {
		t.Fatal("expected epB to be allowed independently of epA")
	}
}

func TestEndpointLimiterWait(t *testing.T) {
	// 10 tokens/sec, burst of 1
	el := NewEndpointLimiter(10.0, 1)
	endpointID := "ep_wait_test"

	// Consume the initial token
	el.Allow(endpointID)

	// Wait should block until token is refilled and succeed within context
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if err := el.Wait(ctx, endpointID); err != nil {
		t.Fatalf("expected Wait to succeed after refill, got error: %v", err)
	}
}
