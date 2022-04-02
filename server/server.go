package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jsirianni/registry/version"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Option is a function that configures a server option
type Option func(*Server)

// WithLogger configures the server's logger
func WithLogger(logger *log.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithWriteTimeout configures the server's write timeout
func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = timeout
	}
}

// WithReadTimeout configures the server's read timeout
func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = timeout
	}
}

// WithListenAddress configures the server's listen address
func WithListenAddress(listenAddr string) Option {
	return func(s *Server) {
		s.listenAddr = listenAddr
	}
}

// Server is the registry web server
type Server struct {
	logger       *log.Logger
	writeTimeout time.Duration
	readTimeout  time.Duration
	listenAddr   string
}

// New takes a logger and returns a new Server
func New(options ...Option) *Server {
	s := &Server{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

// Server starts the web server
func (s *Server) Serve() error {
	r := mux.NewRouter()
	r.HandleFunc("/version", s.versionHandler)
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         s.listenAddr,
		WriteTimeout: s.writeTimeout,
		ReadTimeout:  s.readTimeout,
	}

	return srv.ListenAndServe()
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	v := version.BuildVersion()
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to build version response: %w", err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
}
