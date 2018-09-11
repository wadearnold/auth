// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		bs, err := read(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log("signup", err)
			return
		}

		// read request body
		var signup signupRequest
		if err := json.Unmarshal(bs, &signup); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log("login", err)
			return
		}

		// find user
		u, err := userService.lookupByEmail(signup.Email)
		if err != nil {
			// TODO(adam): should we return the raw error back? info disclosure?
			encodeError(w, err)
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log("signup", err)
			return
		}
		if u == nil {
			var signup signupRequest
			bs, err := ioutil.ReadAll(r.Body)
			if err != nil {
				encodeError(w, err)
				w.WriteHeader(http.StatusBadRequest)
				logger.Log("signup", fmt.Sprintf("failed reading request: %v", err))
				return
			}
			if err := json.Unmarshal(bs, &signup); err != nil {
				encodeError(w, err)
				w.WriteHeader(http.StatusBadRequest)
				logger.Log("signup", fmt.Sprintf("failed parsing request json: %v", err))
				return
			}

			// store user
			userId := generateID()
			if userId == "" {
				encodeError(w, err)
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log("signup", fmt.Sprintf("blank userId generated, err=%v", err))
				return
			}
			u = &User{
				ID:         userId,
				Email:      signup.Email,
				FirstName:  signup.FirstName,
				LastName:   signup.LastName,
				Phone:      signup.Phone,
				CompanyURL: signup.CompanyURL,
				CreatedAt:  time.Now(),
			}
			if err := userService.upsert(u); err != nil {
				// TODO(adam): should we return the raw error back? info disclosure?
				encodeError(w, err)
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log("signup", fmt.Sprintf("problem writing user: %v", err))
				return
			}

			// TODO(adam): check password requirements ?

			if err := auth.write(u.ID, signup.Password); err != nil {
				encodeError(w, errors.New("problem writing user credentials"))
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log("signup", fmt.Sprintf("problem writing user credentials: %v", err))
				return
			}
		} else {
			// user found, so reject signup
			encodeError(w, errors.New("user already exists"))
			w.WriteHeader(http.StatusForbidden)
		}

		// TODO(adam): wipe all old cookies? (ones we got in request)
	}
}
