package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const (
	envSecretKey = "REGISTRY_CONFIG_SECRET_KEY" // #nosec: This is not a credential value
	envStoreType = "REGISTRY_CONFIG_STORAGE_TYPE"
)

func logger() *log.Logger {
	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg:   "message",
			log.FieldKeyLevel: "severity",
		},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.TraceLevel)
	return logger
}

func configure() error {
	providersDir = flag.String("providers-dir", "./providers", "The directory to server providers from")
	certificate = flag.String("certificate", "", "The x509 TLS certificate file (otional)")
	privateKey = flag.String("private-key", "", "The x509 TLS private key file (optional")
	listenPort = flag.Int("port", 8080, "The TCP port to listen on")
	storageType = flag.String("storage-type", "memory", "The storage backend to use")
	// not global, parsed into global uuid
	secretKey := flag.String("secret-key", "", "A UUID secret key, used for authenticating to the server")
	flag.Parse()

	// If not set, check env
	if *secretKey == "" {
		x := os.Getenv(envSecretKey)
		secretKey = &x
	}

	if *secretKey == "" {
		return fmt.Errorf("flag --secret-key or environment %s is a required flag", envSecretKey)
	}

	// If not set, check env
	if *storageType == "" {
		x := os.Getenv(envStoreType)
		storageType = &x
	}

	s, err := uuid.Parse(*secretKey)
	if err != nil {
		return fmt.Errorf("value passed to --secret-key is an invalid UUID: %s", err)
	}
	secretKeyUUID = s

	return nil
}
