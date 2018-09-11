// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2"
)

func main() {
	conf := oauth2.Config{
		ClientID:     "000000",
		ClientSecret: "999999",
		Scopes:       []string{"read"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost:8080/authorize",
			TokenURL: "http://localhost:8080/token",
		},
	}

	// We need to pass along more fields in the token request
	u := "http://localhost:8080/token?grant_type=client_credentials&client_id=000000&client_secret=999999&scope=read"
	resp, err := http.Get(u)
	if err != nil {
		panic(err.Error())
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}

	// Grab (parts) of the token from JSON. Not every field matches up (renewal_token / expiry).
	// TODO(adam): fix? this json read
	var token oauth2.Token
	if err := json.Unmarshal(bs, &token); err != nil {
		panic(err.Error())
	}

	client := conf.Client(context.TODO(), &token)

	u = conf.Endpoint.AuthURL
	resp, err = client.Get(u)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(resp.Status)
}
