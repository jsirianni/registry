package store

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/datastore"
	"github.com/jsirianni/registry/model"
)

// NewDatastore returns a new Google Cloud datastore
func NewDatastore(entityKind string) (*Datastore, error) {
	if entityKind == "" {
		return nil, errors.New("entityKind cannot be empty")
	}

	// TODO: Project is detected at runtime by the Google SDK. This means
	// registry can only run within a GCE / GKE environment. We could expose
	// this parameter to allow registry to run with Cloud Datastore outside
	// of Google Cloud.
	// Settign GOOGLE_APPLICATION_CREDENTIALS is not enough to detect the project.
	projectID := ""

	ctx := context.Background()
	c, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to configure cloud datastore: %s", err)
	}

	return &Datastore{
		kind:   entityKind,
		client: c,
	}, nil
}

// Datastore is a Google Cloud Datastore provider storage backend
type Datastore struct {
	kind   string
	client *datastore.Client
}

var _ Store = &Datastore{}

// Get returns a value from the domain map
func (d *Datastore) Read(name string) (*model.ProviderVersions, error) {
	ctx := context.Background()
	k := datastore.NameKey(d.kind, name, nil)
	e := &model.ProviderVersions{}
	if err := d.client.Get(ctx, k, &e); err != nil {
		return nil, fmt.Errorf("failed to get key %s from cloud datastore: %s", name, err)
	}
	return nil, nil
}

// Set adds a key value pair to the domain map
func (d *Datastore) Write(name string, provider model.ProviderVersions) error {
	ctx := context.Background()
	k := datastore.NameKey(d.kind, name, nil)
	if _, err := d.client.Put(ctx, k, provider); err != nil {
		return fmt.Errorf("failed to create record %s: %s", name, err)
	}
	return nil
}

// Close closes the cloud datastore
func (d *Datastore) Close() error {
	return d.client.Close()
}
