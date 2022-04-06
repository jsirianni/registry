package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jsirianni/registry/model"
	"github.com/jsirianni/registry/store"
	"github.com/jsirianni/registry/version"

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

// WithMapStore configures the server's storage interface with
// an in memory mapstore.
func WithMapStore() Option {
	return func(s *Server) {
		s.store = store.NewMapStore()
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
	store        store.Store
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

// Serve starts the API server
func (s *Server) Serve() error {
	version := version.BuildVersion()

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, version)
	})

	r.GET("/.well-known/terraform.json", func(c *gin.Context) {
		m := map[string]string{
			"providers.v1": "/v1/",
		}
		c.JSON(http.StatusOK, m)
	})

	r.PUT("/v1/:namespace/:name/versions", s.addVersions)
	r.GET("/v1/:namespace/:name/versions", s.getVersions)
	r.GET("/v1/:namespace/:name/:version/download/:os/:arch", s.downloadHandler)

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

// TODO: Require authentication via header key
func (s *Server) addVersions(c *gin.Context) {
	b, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		// TODO: probably check if body is too big
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	var req model.ProviderVersion
	if err := json.Unmarshal(b, &req); err != nil {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	if req.Version == "" {
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf("%s-%s", namespace, name)

	// Check for existing versions
	versions := s.store.Read(key)
	if versions == nil {
		versions = &model.ProviderVersions{}
	}

	// Update and return
	for i, version := range versions.Versions {
		if version.Version == req.Version {
			versions.Versions[i] = req
			s.store.Write(key, *versions)
			c.JSON(http.StatusOK, versions)
			return
		}
	}

	// New resource
	versions.Versions = append(versions.Versions, req)
	s.store.Write(key, *versions)
	c.JSON(http.StatusAccepted, versions)
}

func (s *Server) getVersions(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf("%s-%s", namespace, name)

	providerVersions := s.store.Read(key)
	if providerVersions == nil {
		c.JSON(http.StatusNotFound, nil)
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

	c.JSON(http.StatusOK, response)
}

func (s *Server) downloadHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	key := fmt.Sprintf("%s-%s", namespace, name)

	providerVersions := s.store.Read(key)
	if providerVersions == nil {
		c.JSON(http.StatusNotFound, nil)
		return
	}

	version := c.Param("version")
	os := c.Param("os")
	arch := c.Param("arch")

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
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, response)
}
