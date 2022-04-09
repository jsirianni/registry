package store

import "github.com/jsirianni/registry/model"

// Store provides methods for reading and writing provider
// versions to a storage backend
type Store interface {
	Read(name string) (*model.ProviderVersions, error)
	Write(name string, provider model.ProviderVersions) error
	Close() error
}
