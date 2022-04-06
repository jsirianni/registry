package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jsirianni/registry/server"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

func main() {
	providersDir := flag.String("providers-dir", "./providers", "The directory to server providers from")
	certificate := flag.String("certificate", "", "The x509 TLS certificate file (otional)")
	privateKey := flag.String("private-key", "", "The x509 TLS private key file (optional")
	listenPort := flag.Int("port", 8080, "The TCP port to listen on")
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&log.JSONFormatter{
		FieldMap: log.FieldMap{
			log.FieldKeyMsg:   "message",
			log.FieldKeyLevel: "severity",
		},
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.TraceLevel)

	s := server.New(
		server.WithLogger(logger),
		server.WithProvidersDir(*providersDir),
		server.WithReadTimeout(time.Second*15),
		server.WithWriteTimeout(time.Second*15),
		server.WithListenAddress(fmt.Sprintf(":%d", *listenPort)),
		server.WithTLS(*certificate, *privateKey),
		server.WithMapStore(),
	)

	err := s.Serve()
	if err != nil {
		logger.Fatalf("server exited with error: %s", err)
	}
	logger.Info("server exited cleanly, shutting down")
}
