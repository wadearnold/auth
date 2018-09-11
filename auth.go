// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

type authable interface {
	// invalidate a user's auth (require them to login again)
	invalidate(userId string) error

	// check compares the provided pass for the user.
	// a non-nil error is returned if the passwords don't match
	// or that the userId doesn't exist.
	check(userId string, pass string) error

	// write creates a new auth record for the given id and password
	write(userId string, pass string) error
}

type auth struct{}

func (a *auth) invalidate(userId string) error {
	return nil
}

func (a *auth) check(userId string, pass string) error {
	return nil
}

func (a *auth) write(userId string, pass string) error {
	return nil
}
