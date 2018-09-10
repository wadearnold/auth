// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"time"
)

// purpose
// - handle signup route POST /account
// - dedup users based on "clean_email" (no special chars, i.e. . - + )
// - validate email (by sending with approval_code) [probably diff file]
// - password (bcrypt), per-row salt

// postgres:
//  - accounts (account_id, email, clean_email, password, salt, created_at)
//  - account_approval_codes (account_id, code, valid_until)

type Account struct {
	ID                  string
	FirstName, LastName string
	PhoneNumber         string
	Email               string
	CompanyURL          string
	CreatedAt           time.Time
}
