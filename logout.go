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
		userId, err := auth.findUserId(extractCookie(r).Value)
		if err != nil {
			internalError(w, err, "logout")
			return
		}
		if err := auth.invalidateCookies(userId); err != nil {
			logger.Log("logout", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		authInactivations.With("method", "web").Add(1)
		w.WriteHeader(http.StatusOK)
	}
}
