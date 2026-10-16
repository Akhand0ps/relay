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
	EventTypes      []string       `json:"event_types"`
	Secret          string         `json:"secret"`
	Status          EndpointStatus `json:"status"`
	CreatedAt       time.Time      `json:"created_at"`
}