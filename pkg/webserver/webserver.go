// webserver is a package that provides a simple web server for your
// go application.
// It adds securely relevant headers and generic options to the request.
// It does also support serving a web interface with vite via the frontend package.
package webserver

import (
	"crypto/tls"
	"net/http"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
)

// WebConfig contains the options for the web server
type WebConfig struct {
	// Address to listen for: ":4000"
	Address string

	// If HTTPS should be used, you can configure the path
	// to the certificates (cert.pem and key.pem).
	Certificate Certificate

	// ReadTimeout is the maximum duration for reading the entire request with
	// the body. Specify a negative value to disable the read timeout
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing
	// out writes of the response
	WriteTimeout time.Duration
}

// Certificate contains the paths to the certificates
type Certificate struct {

	// Private Key of the certificate (key.pem)
	PrivateKey string `yaml:"privateKey"`

	// Certificate (cert.pem)
	Certificate string `yaml:"certificate"`
}

type WebServer[T any] struct {

	// Logger to log messages
	Logger *logger.Logger

	// The configuration of the web server
	Config *WebConfig

	// The underlaying http server
	Srv *http.Server

	// Dependency to inject for configureRouter()
	Dependency T
}

// Setup prepares the webserver before starting.
// You have to provide a function which will be called when the
// routes are being set up. It should return a "http.Handler" and does
// receive the webserver with the configured dependency as a parameter.
func (server *WebServer[T]) Setup(configureRouter func(*WebServer[T]) http.Handler) {

	// Apply default values
	if server.Config.ReadTimeout == 0 {
		server.Config.ReadTimeout = 5 * time.Second
	}
	if server.Config.WriteTimeout == 0 {
		server.Config.WriteTimeout = 10 * time.Second
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:       tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	server.Srv = &http.Server{
		Addr:         server.Config.Address,
		ErrorLog:     nil,
		Handler:      configureRouter(server),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  server.Config.ReadTimeout,
		WriteTimeout: server.Config.WriteTimeout,
	}
}

// Start starts the previously configured Webserver.
// The method ListenAndServe blocks until the application dies.
func (server *WebServer[T]) Start() {
	logger.Info("Server started on %s", server.Config.Address)
	var err error
	if server.Config.Certificate.PrivateKey == "" || server.Config.Certificate.Certificate == "" {
		err = server.Srv.ListenAndServe()
	} else {
		err = server.Srv.ListenAndServeTLS(server.Config.Certificate.Certificate, server.Config.Certificate.PrivateKey)
	}
	logger.Error("Failed to run the WebServer: %s", err)
}
