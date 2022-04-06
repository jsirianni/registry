package store

import "github.com/jsirianni/registry/model"

type Store interface {
	Read(name string) *model.ProviderVersions
	Write(name string, provider model.ProviderVersions) error
}
