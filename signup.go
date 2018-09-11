// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	// misc profile information
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone_number"`
	CompanyURL string `json:"company_url"`
}

func addSignupRoutes(router *mux.Router, logger log.Logger, auth authable, userService userRepository) {
	router.Methods("POST").Path("/users/create").HandlerFunc(signupRoute(auth, userService))
}

func signupRoute(auth authable, userService userRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		email := "" // TODO(adam)
		u, _ := userService.lookupByEmail(email)
		// TODO(adam): if user is found error and return back
		if u == nil { // TODO(adam): check err == "no user found"
			var signup signupRequest
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				panic(err.Error())
			}
			if err := json.Unmarshal(bs, &signup); err != nil {
				panic(err.Error())
			}

			u = &User{
				ID:         "", // TODO(adam)
				Email:      signup.Email,
				FirstName:  signup.FirstName,
				LastName:   signup.LastName,
				Phone:      signup.Phone,
				CompanyURL: signup.CompanyURL,
				CreatedAt:  time.Now(),
			}
			if err := userService.upsert(u); err != nil {
				panic(err.Error())
			}

			// TODO(adam): check password requirements ?
			if err := auth.write(u.ID, signup.Password); err != nil {
				panic(err.Error())
			}
		}

		// TODO(adam): wipe all old cookies? (ones we got in request)
		w.WriteHeader(http.StatusOK)
	}
}
