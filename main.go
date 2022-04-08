package main

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jsirianni/registry/server"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func main() {
	providersDir := flag.String("providers-dir", "./providers", "The directory to server providers from")
	certificate := flag.String("certificate", "", "The x509 TLS certificate file (otional)")
	privateKey := flag.String("private-key", "", "The x509 TLS private key file (optional")
	listenPort := flag.Int("port", 8080, "The TCP port to listen on")
	secretKey := flag.String("secret-key", "", "A UUID secret key, used for authenticating to the server")
	flag.Parse()

	logger := log.New()
	logger.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg:   "message",
			log.FieldKeyLevel: "severity",
		},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.TraceLevel)

	const envSecretKey = "REGISTRY_CONFIG_SECRET_KEY" // #nosec: This is not a credential value

	// If not set, check env
	if *secretKey == "" {
		x := os.Getenv(envSecretKey)
		secretKey = &x
	}

	if *secretKey == "" {
		logger.Error("flag --secret-key or environment %s is a required flag", envSecretKey)
		os.Exit(1)
	}

	secretKeyUUID, err := uuid.Parse(*secretKey)
	if err != nil {
		logger.Errorf("value passed to --secret-key is an invalid UUID: %s", err)
	}

	s := server.New(
		server.WithLogger(logger),
		server.WithProvidersDir(*providersDir),
		server.WithReadTimeout(time.Second*15),
		server.WithWriteTimeout(time.Second*15),
		server.WithListenAddress(fmt.Sprintf(":%d", *listenPort)),
		server.WithTLS(*certificate, *privateKey),
		server.WithMapStore(),
		server.WithSecretKey(secretKeyUUID),
	)

	if err := s.Serve(); err != nil {
		logger.Fatalf("server exited with error: %s", err)
	}
	logger.Info("server exited cleanly, shutting down")
}
