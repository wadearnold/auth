// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// purpose
// - handle signup route POST /account
// - dedup users based on "clean_email" (no special chars, i.e. . - + )
// - validate email (by sending with approval_code) [probably diff file]
// - password (bcrypt), per-row salt

// postgres:
//  - users (user_id, email, clean_email, password, salt, created_at)
//  - user_details (user_id, first_name, last_name, company_url)
//  - user_approval_codes (account_id, code, valid_until)

type User struct {
	ID                  string
	Email               string
	FirstName, LastName string
	Phone               string
	CompanyURL          string
	CreatedAt           time.Time
}

var (
	dropPlusExtender = regexp.MustCompile(`(\+.*)$`)
	dropPeriods      = strings.NewReplacer(".", "")
)

func (u *User) cleanEmail() string {
	return cleanEmail(u.Email)
}

// cleanEmail strips all the funky characters from an email address.
//
// Essentially this boils down to the following pattern (ignoring case):
//   [a-z0-9]@[a-z0-9].[a-z]
//
// Callers should be aware of when an empty string is returned.
func cleanEmail(email string) string {
	// split at '@'
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}

	parts[0] = dropPlusExtender.ReplaceAllString(parts[0], "")
	parts[0] = dropPeriods.Replace(parts[0])

	return strings.Join(parts, "@")
}

// generateID creates a new ID for our auth system.
// Do no assume anything about these ID's other than
// they are strings. Case matters
func generateID() string {
	bs := make([]byte, 20)
	n, err := rand.Read(bs)
	if err != nil || n == 0 {
		logger.Log("generateID", fmt.Sprintf("n=%d, err=%v", n, err))
		return ""
	}
	return strings.ToLower(hex.EncodeToString(bs))
}

type userRepository interface {
	lookupById(id string) (*User, error)

	// lookupByEmail finds a user by the given email address.
	// It's recommended you provide User.CleanEmail()
	lookupByEmail(email string) (*User, error)

	upsert(*User) error
}

type sqliteUserRepository struct{}

func (s *sqliteUserRepository) lookupById(id string) (*User, error) {
	return nil, nil
}

func (s *sqliteUserRepository) lookupByEmail(email string) (*User, error) {
	return nil, nil
}

func (s *sqliteUserRepository) upsert(inc *User) error {
	return nil
}
