// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// maxReadBytes is the number of bytes to read
	// from a request body. It's intended to be used
	// with an io.LimitReader
	maxReadBytes = 1 * 1024 * 1024

	cookieName = "moov-io--auth"
	cookieTTL  = 30 * 24 * time.Hour // days * hours/day * hours
)

// read consumes an io.Reader (wrapping with io.LimitReader)
// and returns either the resulting bytes or a non-nil error.
func read(r io.Reader) ([]byte, error) {
	r = io.LimitReader(r, maxReadBytes)
	return ioutil.ReadAll(r)
}

// encodeError JSON encodes the supplied error
//
// The HTTP status of "400 Bad Request" is written to the
// response.
func encodeError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func internalError(w http.ResponseWriter, err error, component string) {
	internalServerErrors.Add(1)
	w.WriteHeader(http.StatusInternalServerError)
	logger.Log(component, err)
}

// extractCookie attempts to pull out our cookie from the incoming request.
// We use the contents to find the associated userId.
func extractCookie(r *http.Request) *http.Cookie {
	if r == nil {
		return nil
	}
	cs := r.Cookies()
	for i := range cs {
		if cs[i].Name == cookieName {
			return cs[i]
		}
	}
	return nil
}

// createCookie generates a new cookie and associates it with the provided
// userId.
func createCookie(userId string, auth authable) (*http.Cookie, error) {
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    generateID(),
		Expires:  time.Now().Add(cookieTTL),
		Secure:   true,
		HttpOnly: true,
	}
	if err := auth.writeCookie(userId, cookie); err != nil {
		return nil, err
	}
	return cookie, nil
}
