package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jsirianni/registry/model"
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
			"providers.v1": "/terraform/providers/v1/",
		}
		c.JSON(http.StatusOK, m)
	})

	r.GET("/terraform/providers/v1/:namespace/:name/versions", s.versionsHandler)

	r.GET("/terraform/providers/v1/:namespace/:name/:version/download/:os/:arch", s.downloadHandler)

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

func (s *Server) versionsHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	path := filepath.Join(s.providersDir, namespace, fmt.Sprintf("%s.json", name))
	fileBytes, err := ioutil.ReadFile(path) // #nosec, used defined relative path based on url params
	if err != nil {
		c.Status(http.StatusNotFound)
		s.logger.Errorf("failed to open file for namespace %s and name %s: %s", namespace, name, err)
		return
	}

	var providerVersions model.ProviderVersions
	if err := json.Unmarshal(fileBytes, &providerVersions); err != nil {
		c.Status(http.StatusInternalServerError)
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

	c.JSON(http.StatusOK, response)
}

func (s *Server) downloadHandler(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	version := c.Param("version")
	os := c.Param("os")
	arch := c.Param("arch")

	path := filepath.Join(s.providersDir, namespace, fmt.Sprintf("%s.json", name))
	fileBytes, err := ioutil.ReadFile(path) // #nosec, used defined relative path based on url params
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	var providerVersions model.ProviderVersions
	if err := json.Unmarshal(fileBytes, &providerVersions); err != nil {
		c.Status(http.StatusInternalServerError)
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
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, response)
}
