package store

import (
	"sync"

	"github.com/jsirianni/registry/model"
)

// NewMap returns a new memory backed storage
func NewMap() *Map {
	return &Map{
		providers: make(map[string]model.ProviderVersions),
	}
}

// Map is an in memory provider storage backend
type Map struct {
	providers map[string]model.ProviderVersions
	mu        sync.RWMutex
}

var _ Store = &Map{}

// Read reads a provider
func (m *Map) Read(name string) (*model.ProviderVersions, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	x, ok := m.providers[name]
	if !ok {
		return nil, nil
	}
	return &x, nil
}

// Write saves a provider
func (m *Map) Write(name string, provider model.ProviderVersions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider

	return nil
}

// Close is a noop
func (m *Map) Close() error {
	return nil
}
