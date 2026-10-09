package models


import (
	"encoding/json"
	"time"
)


type EventStatus string

const (

	StatusQueued EventStatus = "queued"
	StatusProcessing EventStatus = "processing"
	StatusSuccess EventStatus = "success"
	StatusFailed EventStatus = "failed"
)

type Event struct {
	ID             string           `json:"id"`
	TenantID       string           `json:"tenant_id,omitempty"`
	Type           string           `json:"type"`
	Payload        json.RawMessage  `json:"payload"`
	IdempotencyKey string           `json:"idempotency_key.omitempty"`
	Status         EventStatus      `json:"status"`
	CreatedAt      time.Time        `json:"created_at"`
}


