// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addLogoutRoutes(router *mux.Router, logger log.Logger, auth authable) {
	router.Methods("DELETE").Path("/users/login").HandlerFunc(logoutRoute(auth))
}

func logoutRoute(auth authable) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO(adam): get u.ID from request (cookie)
		// TODO(adam): that extraction will be used in all other routes
		id := "" // TODO
		if err := auth.invalidate(id); err != nil {
			// TODO(adam): log or metrics
		}
		authInactivations.With("method", "web").Add(1)
		w.WriteHeader(http.StatusOK)
	}
}
