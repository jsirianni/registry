package store

import (
	"sync"

	"github.com/jsirianni/registry/model"
)

func NewMapStore() *Map {
	return &Map{
		providers: make(map[string]model.ProviderVersions),
	}
}

type Map struct {
	providers map[string]model.ProviderVersions
	mu        sync.RWMutex
}

var _ Store = &Map{}

func (m *Map) Read(name string) *model.ProviderVersions {
	m.mu.RLock()
	defer m.mu.RUnlock()

	x, ok := m.providers[name]
	if !ok {
		return nil
	}
	return &x
}

func (m *Map) Write(name string, provider model.ProviderVersions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[name] = provider

	return nil
}
