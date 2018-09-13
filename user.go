// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
)

// purpose
// - handle signup route POST /account
// - dedup users based on "clean_email" (no special chars, i.e. . - + )
// - validate email (by sending with approval_code) [probably diff file]

// TODO(adam): having a row in user_approval_codes means user isn't verified. Delete on email approval.

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
	// It's recommended you provide User.cleanEmail()
	lookupByEmail(email string) (*User, error)

	upsert(*User) error
}

type sqliteUserRepository struct {
	db  *sql.DB
	log log.Logger
}

func (s *sqliteUserRepository) lookupById(id string) (*User, error) {
	// users and user_details
	return nil, nil
}

func (s *sqliteUserRepository) lookupByEmail(email string) (*User, error) {
	// users
	var u *User
	return s.lookupById(u.ID)
}

func (s *sqliteUserRepository) upsert(inc *User) error {
	// users and user_details
	return nil
}

// authable represents the interactions of a user's authentication
// status. This boils down to password comparison and cookie data.
type authable interface {
	findUserId(data string) (string, error)
	invalidateCookies(userId string) error
	writeCookie(userId string, incoming string) error

	// checkPassword compares the provided password for the user.
	// a non-nil error is returned if the passwords don't match
	// or that the userId doesn't exist.
	checkPassword(userId string, pass string) error
	writePassword(userId string, pass string) error
}

type auth struct {
	db  *sql.DB
	log log.Logger
}

// findUserId takes cookie data and returns the userId associated
func (a *auth) findUserId(data string) (string, error) {
	query := `select user_id from user_cookies where data == ? and valid_until > ?`
	stmt, err := a.db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	rows, err := stmt.Query(data)
	if err != nil {
		return "", err
	}
	var userId string
	for rows.Next() {
		rows.Scan(&userId)
		if userId != "" {
			return userId, nil
		}
	}
	return "", nil
}

func (a *auth) invalidateCookies(userId string) error {
	// user_cookies
	return nil
}

func (a *auth) writeCookie(userId string, data string) error {
	// user_cookies
	return nil
}

func (a *auth) checkPassword(userId string, pass string) error {
	// user_passwords
	return nil
}

func (a *auth) writePassword(userId string, pass string) error {
	// user_passwords
	return nil
}
