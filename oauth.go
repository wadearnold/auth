// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/moov-io/auth/pkg/buntdbclient"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"gopkg.in/oauth2.v3"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

type oauth struct {
	manager     *manage.Manager
	clientStore *oauth2.ClientStore
	server      *server.Server

	logger log.Logger
}

func setupOauthServer(logger log.Logger) (*oauth, error) {
	out := &oauth{
		logger: logger,
	}

	// oauth2 setup
	path := os.Getenv("OAUTH2_TOKENS_DB_PATH")
	if path == "" {
		path = "oauth2_tokens.db"
	}
	tokenStore, err := store.NewFileTokenStore(path)
	if err != nil {
		return nil, fmt.Errorf("problem creating token store: %v", err)
	}

	out.manager = manage.NewDefaultManager()
	out.manager.MapTokenStorage(tokenStore)

	path = os.Getenv("OAUTH2_CLIENTS_DB_PATH")
	if path == "" {
		path = "oauth2_clients.db"
	}
	cs, err := buntdbclient.New(path)
	if err != nil {
		return nil, fmt.Errorf("problem creating clients store: %v", err)
	}
	out.clientStore = &cs
	out.manager.MapClientStorage(*out.clientStore)

	out.server = server.NewDefaultServer(out.manager)
	out.server.SetAllowGetAccessRequest(true)
	out.server.SetClientInfoHandler(server.ClientFormHandler)
	out.server.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		logger.Log("internal-error", err.Error())
		return
	})
	out.server.SetResponseErrorHandler(func(re *errors.Response) {
		logger.Log("response-error", re.Error.Error())
		return
	})

	return out, nil
}

// addOAuthRoutes includes our oauth2 routes on the provided mux.Router
func addOAuthRoutes(r *mux.Router, o *oauth, logger log.Logger) {
	r.Methods("GET").Path("/authorize").HandlerFunc(o.authorizeHandler)
	if o.server.Config.AllowGetAccessRequest {
		r.Methods("GET").Path("/token").HandlerFunc(o.tokenHandler)
	} else {
		r.Methods("POST").Path("/token").HandlerFunc(o.tokenHandler)
	}
}

// authorizeHandler checks the request for appropriate oauth information
// and returns "200 OK" if the token is valid.
func (o *oauth) authorizeHandler(w http.ResponseWriter, r *http.Request) {
	// We aren't using HandleAuthorizeRequest here because that assumes redirect_uri
	// exists on the request. We're just checking for a valid token.
	ti, err := o.server.ValidationBearerToken(r)
	if err != nil {
		authFailures.With("method", "oauth2").Add(1)
		encodeError(w, err)
		return
	}
	if ti.GetClientID() == "" {
		authFailures.With("method", "oauth2").Add(1)
		encodeError(w, fmt.Errorf("missing client_id"))
		return
	}

	// Passed token check, return "200 OK"
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	authSuccesses.With("method", "oauth2").Add(1)
}

// tokenHandler passes off the request down to our oauth2 library to
// generate a token (or return an error).e
func (o *oauth) tokenHandler(w http.ResponseWriter, r *http.Request) {
	err := o.server.HandleTokenRequest(w, r)
	if err != nil {
		encodeError(w, err)
		return
	}
	// TODO(adam): We need to track this metric inside our TokenStorage.
	// HandleTokenRequest currently returns nil even if the token request
	// failed. There's no real way to inspect the http.ResponseWriter in
	// an attempt to correctly calculate this.
	// tokenGenerations.With("method", "oauth2").Add(1)
}
