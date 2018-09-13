// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupServer() *Server {
	timeout, _ := time.ParseDuration("45s")
	return &Server{
		svc: &http.Server{
			Addr:         ":9090",
			Handler:      handler(),
			ReadTimeout:  timeout,
			WriteTimeout: timeout,
			IdleTimeout:  timeout,
		},
	}
}

// Server represents a holder around a net/http Server which
// is used for admin endpoints. (i.e. metrics, healthcheck)
type Server struct {
	svc *http.Server
}

func (s *Server) BindAddress() string {
	return s.svc.Addr
}

// Start brings up the admin HTTP service. This call blocks.
func (s *Server) Listen() error {
	if s == nil || s.svc == nil {
		return nil
	}
	return s.svc.ListenAndServe()
}

// Shutdown unbinds the HTTP server.
func (s *Server) Shutdown() {
	if s == nil || s.svc == nil {
		return
	}
	s.svc.Shutdown(nil)
}

// pprofHandlers is a map which holds configuration for
// serving various pprof handlers on the admin servlet.
//
// We only wnat to expose on the admin servlet because these
// profiles/dumps can contain sensitive info (i.e. password hash, email)
// or alter the app performance.
//
// These values can be overridden by setting the appropiate PPROF_*
// environment variable.
//
// TODO(adam): generate readme.md env vars from this
var pprofHandlers = map[string]bool{
	"allocs":       true,
	"block":        true,
	"cmdline":      true,
	"goroutine":    true,
	"heap":         true,
	"mutex":        true,
	"profile":      true,
	"threadcreate": false,
	"trace":        false,
}

// checkEnabled calls out to os.Getenv with the following pattern:
//  PPROF_$name where $name is uppercase
//
// A string of "yes" returns true, and "no" returns false, otherwise
// zero is returned.
func checkEnabled(name string, zero bool) bool {
	v := os.Getenv(fmt.Sprintf("PPROF_%s", strings.ToUpper(name)))
	switch strings.ToLower(v) {
	case "yes":
		return true
	case "no":
		return false
	}
	return zero
}

func handler() http.Handler {
	r := mux.NewRouter()

	// prometheus metrics
	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	// add all pprof handlers we've configured
	r.HandleFunc("/debug/pprof/", pprof.Index)
	for k, add := range pprofHandlers {
		if checkEnabled(k, add) {
			r.Handle(fmt.Sprintf("/debug/pprof/%s", k), pprof.Handler(k))
		}
	}

	return r
}
