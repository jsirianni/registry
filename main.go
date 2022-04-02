package main

import (
	"os"
	"time"

	"github.com/jsirianni/registry/server"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.TraceLevel)

	s := server.New(
		server.WithLogger(logger),
		server.WithReadTimeout(time.Second*15),
		server.WithWriteTimeout(time.Second*15),
		server.WithListenAddress(":8000"),
	)

	logger.Info("starting server")
	err := s.Serve()
	if err != nil {
		logger.Fatalf("server exited with error: %w", err)
	}
	logger.Info("server exited cleanly, shutting down")
}
