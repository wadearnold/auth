// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`

	// misc profile information
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Phone      string `json:"phone"`
	CompanyURL string `json:"companyUrl,omitempty"`
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
			internalError(w, err, "signup")
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
		if err != nil && !strings.Contains(err.Error(), "user not found") {
			// TODO(adam): should we return the raw error back? info disclosure?
			encodeError(w, err)
			internalError(w, err, "signup")
			return
		}
		if u == nil {
			var signup signupRequest
			if err := json.Unmarshal(bs, &signup); err != nil {
				encodeError(w, err)
				logger.Log("signup", fmt.Sprintf("failed parsing request json: %v", err))
				return
			}

			// store user
			userId := generateID()
			if userId == "" {
				encodeError(w, err)
				internalError(w, fmt.Errorf("blank userId generated, err=%v", err), "signup")
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
				internalError(w, fmt.Errorf("problem writing user: %v", err), "signup")
				return
			}

			if err := auth.writePassword(u.ID, signup.Password); err != nil {
				encodeError(w, errors.New("problem writing user credentials"))
				internalError(w, fmt.Errorf("problem writing user credentials: %v", err), "signup")
				return
			}

			// TODO(adam): signup worked, so render back user and oauth client info
			//
			// On signup, we create an oauth2 (model.Token) with random client id/secret,
			// domain (todo?), and UserID set to our value, write that (using o.clientStore).

		} else {
			// user found, so reject signup
			encodeError(w, errors.New("user already exists"))
		}
	}
}
