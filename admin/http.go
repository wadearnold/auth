// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package admin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
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
	s.svc.Shutdown(context.TODO())
}

func handler() http.Handler {
	r := mux.NewRouter()

	// prometheus metrics
	r.Methods("GET").Path("/metrics").Handler(promhttp.Handler())

	// add all pprof handlers we've configured
	r.HandleFunc("/debug/pprof/", pprof.Index)
	for k, add := range pprofHandlers {
		if pprofProfileEnabled(k, add) {
			r.Handle(fmt.Sprintf("/debug/pprof/%s", k), pprof.Handler(k))
		}
	}

	return r
}
