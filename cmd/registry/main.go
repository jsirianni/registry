package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jsirianni/registry/server"
)

// globals set by configure()
var (
	providersDir  *string
	certificate   *string
	privateKey    *string
	storageType   *string
	listenPort    *int
	secretKeyUUID uuid.UUID
)

func main() {
	logger := logger()

	if err := configure(); err != nil {
		logger.Errorf("configure error: %v", err)
		os.Exit(1)
	}

	var storeOption server.Option
	switch t := *storageType; t {
	case "memory":
		storeOption = server.WithMapStore()
	case "datastore":
		const kind = "tfregistry" // TODO: This should probably be configurable
		storeOption = server.WithCloudDatastore(kind)
	default:
		logger.Errorf("invalid storage type '%s'", t)
		os.Exit(1)
	}

	s, err := server.New(
		server.WithLogger(logger),
		server.WithProvidersDir(*providersDir),
		server.WithReadTimeout(time.Second*15),
		server.WithWriteTimeout(time.Second*15),
		server.WithListenAddress(fmt.Sprintf(":%d", *listenPort)),
		server.WithTLS(*certificate, *privateKey),
		server.WithSecretKey(secretKeyUUID),
		storeOption,
	)
	if err != nil {
		logger.Errorf("server configuration: %v", err)
		os.Exit(1)
	}

	if err := s.Serve(); err != nil {
		logger.Errorf("server exited with error: %s", err)
		os.Exit(1)
	}
	logger.Info("server exited cleanly, shutting down")
}
