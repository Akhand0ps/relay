package store

import "github.com/akhand0ps/relay/internal/models"



type Store interface {

	Save(event models.Event) error
	Get(id string) (*models.Event,error)
	SaveEndpoint(ep models.Endpoint)error
	GetEndpoint(id string) (*models.Endpoint,error)
	ListEndpoint()([]models.Endpoint,error)
	FindEndPointForEvent(eventType string) ([]models.Endpoint,error)
	GetByIdempotencyKey(key string)(*models.Event,bool)  
}