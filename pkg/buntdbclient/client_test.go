// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package buntdbclient

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/oauth2.v3/models"
)

func TestClientStore(t *testing.T) {
	dir, err := ioutil.TempDir("", "moov-auth-client")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db, _ := New(filepath.Join(dir, "client_test.db"))
	cs, _ := db.(*ClientStore)

	id := "moov"

	// get nothing
	cli, err := cs.GetByID(id)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("got %#v", err)
	}
	if cli.GetID() != "" {
		t.Errorf("got %#v", err)
	}

	// set something
	err = cs.Set(id, &models.Client{
		ID:     id,
		Secret: "secret",
		Domain: "domain",
		UserID: "userId",
	})
	if err != nil {
		t.Errorf("got %v", err)
	}

	// get something
	cli, err = cs.GetByID(id)
	if err != nil {
		t.Errorf("got %v", err)
	}
	if cli.GetID() != id {
		t.Errorf("got %s", cli.GetID())
	}
	if cli.GetSecret() != "secret" {
		t.Errorf("got %s", cli.GetSecret())
	}
	if cli.GetDomain() != "domain" {
		t.Errorf("got %s", cli.GetDomain())
	}
	if cli.GetUserID() != "userId" {
		t.Errorf("got %s", cli.GetUserID())
	}
}
