package store

import (
	"errors"
	"sync"

	"github.com/akhand0ps/relay/internal/models"
)

type MemoryStore struct {
	events map[string]models.Event
	mu  sync.RWMutex
}


func NewMemoryStore() *MemoryStore{

	return &MemoryStore{
		events: make(map[string]models.Event),
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
