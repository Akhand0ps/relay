package store

import (
	"errors"
	"sync"

	"github.com/akhand0ps/relay/internal/models"
)

type MemoryStore struct {
	events map[string]models.Event
	mu  sync.RWMutex
	endpoints map[string]models.Endpoint
}


func NewMemoryStore() *MemoryStore{

	return &MemoryStore{
		events: make(map[string]models.Event),
		endpoints: make(map[string]models.Endpoint),
	}
}


func (m *MemoryStore) Save(event models.Event) error {

	m.mu.Lock()
	defer m.mu.Unlock()
	m.events[event.ID] = event
	return nil
}

func (m *MemoryStore) Get(id string) (*models.Event, error) {


	m.mu.RLock()
	defer m.mu.RUnlock()

	val,prs := m.events[id]

	if !prs {
		return nil, errors.New("Event not FOUND")
	}
	return &val,nil
}

func (m *MemoryStore) SaveEndpoint(ep models.Endpoint) error {

	m.mu.Lock()
	defer m.mu.Unlock()
	m.endpoints[ep.ID] = ep
	return nil
}


func (m * MemoryStore) GetEndpoint(id string) (*models.Endpoint,error){

	m.mu.RLock()
	defer m.mu.RUnlock()

	val,prs := m.endpoints[id]

	if !prs {
		return nil,errors.New("Endpoinnt not found")
	}

	return &val,nil
}

func (m *MemoryStore)ListEndpoint()([]models.Endpoint,error){

	m.mu.RLock()
	defer m.mu.RUnlock()

	list:= make([]models.Endpoint,0)
	for _,ep:=range m.endpoints {
		list=append(list,ep)
	}
	return list,nil
}


func (m * MemoryStore) FindEndPointForEvent(eventType string) ([]models.Endpoint,error){

	m.mu.RLock()
	defer m.mu.RUnlock()

	endpoints := make([]models.Endpoint,0)

	for _,ep:= range m.endpoints {
		if ep.Status == models.EndpointStatusActive {
			//check if eventtype is in ep.eventTypes
			for _,etype:=range ep.EventTypes {
				if etype == eventType {
					endpoints = append(endpoints,ep)
					break
				}
			}
		}
	}
	return endpoints,nil
}



