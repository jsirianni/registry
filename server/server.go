package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jsirianni/registry/model"
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

// WithProvidersDir configures the server's logger
func WithProvidersDir(dir string) Option {
	return func(s *Server) {
		s.providersDir = dir
	}
}

// WithTLS configures the server's optional TLS config, panics
// if tls.LoadX509KeyPair fails.
func WithTLS(crt, key string) Option {
	if crt == "" || key == "" {
		return nil
	}

	return func(s *Server) {
		c, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			panic(fmt.Sprintf("failed to load tls keypair: %s", err))
		}

		s.tls = &tls.Config{
			Certificates: []tls.Certificate{c},
			MinVersion:   tls.VersionTLS12,
		}
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
	providersDir string
	tls          *tls.Config
}

// New takes a logger and returns a new Server
func New(options ...Option) *Server {
	s := &Server{}
	for _, opt := range options {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// Serve starts the web server
func (s *Server) Serve() error {
	r := mux.NewRouter()

	// health endpoint for healthchecks
	r.HandleFunc("/health", s.healthHandler).Methods(http.MethodGet)

	// version endpoint returns the server's verion and build info
	r.HandleFunc("/version", s.versionHandler).Methods(http.MethodGet)

	// returns the server's supported resources
	r.HandleFunc("/.well-known/terraform.json",
		s.discoverHandler).Methods(http.MethodGet)

	// returns a given provider's versions
	r.HandleFunc("/terraform/providers/v1/{namespace}/{name}/versions",
		s.versionsHandler).Methods(http.MethodGet)

	// returns information on how to download a given version
	r.HandleFunc(
		"/terraform/providers/v1/{namespace}/{name}/{version}/download/{os}/{arch}",
		s.downloadHandler).Methods(http.MethodGet)

	//r.PathPrefix("/").HandlerFunc(s.catchAllHandler)

	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         s.listenAddr,
		WriteTimeout: s.writeTimeout,
		ReadTimeout:  s.readTimeout,
	}

	if s.tls != nil {
		srv.TLSConfig = s.tls
		s.logger.Debugf("starting server with TLS")
		return srv.ListenAndServeTLS("", "")
	}
	return srv.ListenAndServe()
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Validate that the server is happy and return 503 if not.
	w.WriteHeader(http.StatusOK)
	s.logger.Tracef("%d %s", http.StatusOK, r.URL.String())
}

func (s *Server) versionHandler(w http.ResponseWriter, r *http.Request) {
	v := version.BuildVersion()
	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to build version response: %s", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
	s.logger.Tracef("%d %s", http.StatusOK, r.URL.String())
}

func (s *Server) discoverHandler(w http.ResponseWriter, r *http.Request) {
	type discovery struct {
		Providers string `json:"providers.v1"`
	}

	d := discovery{
		Providers: "/terraform/providers/v1/",
	}

	b, err := json.Marshal(d)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to build discovery response: %s", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
	s.logger.Tracef("%d %s", http.StatusOK, r.URL.String())
}

func (s *Server) versionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	namespace, ok := vars["namespace"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Debugf("namespace not set %s", r.URL.String())
		return
	}

	name, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		s.logger.Debugf("name not set %s", r.URL.String())
		return
	}

	path := filepath.Join(s.providersDir, namespace, fmt.Sprintf("%s.json", name))
	fileBytes, err := ioutil.ReadFile(path) // #nosec, used defined relative path based on url params
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Errorf("failed to open file for namespace %s and name %s: %s", namespace, name, err)
		return
	}

	var providerVersions model.ProviderVersions
	if err := json.Unmarshal(fileBytes, &providerVersions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to unmarhsal %s into type model.ProviderVersions: %s", path, err)
		return
	}

	versions := []model.Version{}
	for _, v := range providerVersions.Versions {
		version := model.Version{
			Version:   v.Version,
			Protocols: v.Protocols,
		}

		for _, p := range v.Platforms {
			// TODO: model.Version should probably break out the Platforms field
			// into its own type so appending is not so gross.
			version.Platforms = append(version.Platforms, struct {
				Os   string "json:\"os\""
				Arch string "json:\"arch\""
			}{p.Os, p.Arch})
		}

		versions = append(versions, version)
	}

	response := make(map[string][]model.Version)
	response["versions"] = versions

	b, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to marshal response: %s", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
	s.logger.Tracef("%d %s", http.StatusOK, r.URL.String())
}

func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	for _, expect := range []string{"namespace", "name", "version", "os", "arch"} {
		_, ok := vars[expect]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			s.logger.Debugf("%s not set %s", expect, r.URL.String())
			return
		}
	}

	namespace := vars["namespace"]
	name := vars["name"]
	version := vars["version"]
	os := vars["os"]
	arch := vars["arch"]

	path := filepath.Join(s.providersDir, namespace, fmt.Sprintf("%s.json", name))
	fileBytes, err := ioutil.ReadFile(path) // #nosec, used defined relative path based on url params
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Errorf("failed to open file for namespace %s and name %s: %s", namespace, name, err)
		return
	}

	var providerVersions model.ProviderVersions
	if err := json.Unmarshal(fileBytes, &providerVersions); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to unmarhsal %s into type model.ProviderVersions: %s", path, err)
		return
	}

	providerVersion := model.ProviderVersion{}
	for _, v := range providerVersions.Versions {
		if v.Version == version {
			providerVersion = v
		}
	}

	response := model.DownloadResponse{
		Protocols: providerVersion.Protocols,
	}

	// check os
	found := false
	for _, x := range providerVersion.Platforms {
		if x.Os == os && x.Arch == arch {
			response.Os = x.Os
			response.Arch = x.Arch
			response.Filename = x.Filename
			response.DownloadURL = x.DownloadURL
			response.ShasumsURL = x.ShasumsURL
			response.ShasumsSignatureURL = x.ShasumsSignatureURL
			response.Shasum = x.Shasum
			response.SigningKeys = x.SigningKeys

			found = true
		}
	}

	if !found {
		w.WriteHeader(http.StatusNotFound)
		s.logger.Debugf("failed to find provider %s for os %s and arch %s", name, os, arch)
		return
	}

	b, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Errorf("failed to marshal response: %s", err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(b))
	s.logger.Tracef("%d %s", http.StatusOK, r.URL.String())
}

func (s *Server) catchAllHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	s.logger.Tracef("%d %s", http.StatusNotImplemented, r.URL.String())
}
