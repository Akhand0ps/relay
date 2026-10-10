package models

import "time"




type EndpointStatus string 

const (

	EndpointStatusActive EndpointStatus = "active"
	EndpointStatusINactive EndpointStatus = "inactive"

	
)


type Endpoint struct {

	ID              string         `json:"id"`
	URL             string         `json:"url"`
	EventID         []string       `json:"event_ids"`
	Status          EndpointStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`	
}