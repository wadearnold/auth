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
	"unicode/utf8"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

const (
	minPasswordLength = 8
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
			encodeError(w, errors.New("if this user exists, please try again with proper credentials"))
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

			// Basic data sanity checks
			if signup.Email == "" {
				encodeError(w, errors.New("no email provided"))
				return
			}
			if signup.Password == "" {
				encodeError(w, errors.New("no password provided"))
				return
			}
			if n := utf8.RuneCountInString(signup.Password); n < minPasswordLength {
				encodeError(w, fmt.Errorf("password required to be at least %d characters", n))
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
				internalError(w, fmt.Errorf("problem writing user: %v", err), "signup")
				return
			}

			if err := auth.writePassword(u.ID, signup.Password); err != nil {
				internalError(w, fmt.Errorf("problem writing user credentials: %v", err), "signup")
				return
			}

			// signup worked, yay!
		} else {
			// user found, so reject signup
			encodeError(w, errors.New("user already exists"))
		}
	}
}
