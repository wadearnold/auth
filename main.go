// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/moov-io/auth/admin"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

var (
	httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")

	logger log.Logger

	// Metrics
	authSuccesses = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: "auth_successes",
		Help: "Count of successful authorizations",
	}, []string{"method"})
	authFailures = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: "auth_failures",
		Help: "Count of failed authorizations",
	}, []string{"method"})

	tokenGenerations = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Name: "auth_token_generations",
		Help: "Count of auth tokens created",
	}, []string{"method"})
)

const Version = "0.1.0-dev"

func main() {
	flag.Parse()

	// Setup logging, default to stdout
	logger = log.NewLogfmtLogger(os.Stderr)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)
	logger.Log("startup", fmt.Sprintf("Starting auth server version %s", Version))

	// Listen for application termination.
	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	oauth, err := setupOauthServer(logger)
	if err != nil {
		panic(err.Error()) // TODO(adam)
	}
	handler := oauthHandler(oauth, logger)

	readTimeout, _ := time.ParseDuration("30s")
	writTimeout, _ := time.ParseDuration("30s")
	idleTimeout, _ := time.ParseDuration("60s")

	serve := &http.Server{
		Addr:    *httpAddr,
		Handler: handler,
		TLSConfig: &tls.Config{
			InsecureSkipVerify:       false,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
		ReadTimeout:  readTimeout,
		WriteTimeout: writTimeout,
		IdleTimeout:  idleTimeout,
	}
	shutdownServer := func() {
		if err := serve.Shutdown(nil); err != nil {
			logger.Log("shutdown", err)
		}
	}

	adminService := admin.SetupServer()
	go func() {
		logger.Log("admin", fmt.Sprintf("Starting admin service on %s", adminService.BindAddress()))
		if err := adminService.Listen(); err != nil {
			logger.Log("admin", "shutting down", "error", err)
		}
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- serve.ListenAndServe()
		// TODO(adam): support TLS
		// func (srv *Server) ListenAndServeTLS(certFile, keyFile string) error
	}()

	if err := <-errs; err != nil {
		adminService.Shutdown()
		shutdownServer()
		logger.Log("exit", err)
	}
}
