// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package main

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type loginRequest struct {
	email    string `json:"email"`
	password string `json:"password"`
}

func addLoginRoutes(router *mux.Router, logger log.Logger, auther authable, userService userRepository) {
	router.Methods("POST").Path("/users/login").HandlerFunc(loginRoute(auther, userService))
}

func loginRoute(auther authable, userService userRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := "" // TODO(adam)
		u, err := userService.lookupByEmail(email)
		if err != nil {
			panic(err.Error())
		}

		pass := "" // TOOD(adam)
		if err := auther.check(u.ID, pass); err != nil {
			w.WriteHeader(http.StatusForbidden)

		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
