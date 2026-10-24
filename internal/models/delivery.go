
package models

import (
	"time"
)


type DeliveryStatus string

const (

	DeliveryStatusQueued DeliveryStatus = "queued"
	DeliveryStatusProcessing DeliveryStatus = "processing"
	DeliveryStatusSuccess DeliveryStatus = "success"
	DeliveryStatusFailed DeliveryStatus = "failed"
	DeliveryStatusRetrying DeliveryStatus = "retrying"
	DeliveryStatusDLQ      DeliveryStatus =  "dlq"
)

type Delivery struct {

	ID                  string               `json:"id"`
	EventId             string               `json:"event_id"`
	EndpointId          string               `json:"endpoint_id"`
	Status              DeliveryStatus       `json:"status"`
	AttemptCount        int                  `json:"attempts"`
	NextAttemptAt       *time.Time           `json:"next_attempt_at"`
	LastAttemptAt       *time.Time           `json:"last_attempt_at"`
	CreatedAt           time.Time            `json:"created_at"`
	UpdatedAt           time.Time            `json:"updated_at"`
}

type DeliveryTask struct {

	Delivery Delivery
	Endpoint Endpoint
	Event    Event
}

type DeliveryAttempt struct {
	ID     string    `json:"id"`
	DeliveryId string  `json:"deliver_id"`
	StatusCode  int  `json:"status_code"`
	ErrorMessage string `json:"error_message,omitempty"`
	ExecutionDurationMs int `json:"execution_duration_ms"`
	AttemptedAt   time.Time  `json:"attempted_at"`
}