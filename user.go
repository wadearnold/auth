// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"golang.org/x/crypto/bcrypt"
)

// purpose
// - handle signup route POST /account
// - dedup users based on "clean_email" (no special chars, i.e. . - + )
// - validate email (by sending with approval_code) [probably diff file]

// TODO(adam): having a row in user_approval_codes means user isn't verified. Delete on email approval.

type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Phone      string    `json:"phone"`
	CompanyURL string    `json:"companyUrl"`
	CreatedAt  time.Time `json:"createdAt"`
}

var (
	dropPlusExtender = regexp.MustCompile(`(\+.*)$`)
	dropPeriods      = strings.NewReplacer(".", "")
)

const (
	bcryptCostFactor = 10 // TODO(adam): value ok?

	// from 'go doc time Time.String'
	serializedTimestampFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
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
	writeCookie(userId string, cookie *http.Cookie) error

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
	// the SHA256 checksum is stored, not the actual data.
	data, err := hash(data)
	if err != nil {
		return "", err
	}

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
	defer rows.Close()
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
	stmt, err := a.db.Prepare(`delete from user_cookies where user_id = ?`)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(userId)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	a.log.Log("user", fmt.Sprintf("deleted %d cookies for userId=%s", n, userId))
	return nil
}

func (a *auth) writeCookie(userId string, cookie *http.Cookie) error {
	stmt, err := a.db.Prepare(`insert into user_cookies values (?, ?, ?)`)
	if err != nil {
		return err
	}

	// hash the data
	data, err := hash(cookie.Value)
	if err != nil {
		return err
	}
	validUntil := cookie.Expires.Format(serializedTimestampFormat)

	// write row
	stmt.Exec(userId, data, validUntil)
	return nil
}

// fakeBcryptRounds just performs a bcrypt.GenerateFromPassword and then
// a bcrypto.CompareHashAndPassword afterwords. In an attempt to make happy
// and sad paths take "approximately" the same.
func (a *auth) fakeBcryptRounds() {
	id := generateID()
	bcrypt.GenerateFromPassword([]byte(id), bcryptCostFactor)
	bcrypt.CompareHashAndPassword([]byte(id), []byte(id))
}

func (a *auth) checkPassword(userId string, incoming string) error {
	stmt, err := a.db.Prepare(`select password, salt from user_passwords where user_id = ?`)
	if err != nil {
		a.fakeBcryptRounds()
		return err
	}

	row := stmt.QueryRow(userId)
	var storedPassword, storedSalt string
	if err := row.Scan(&storedPassword, storedSalt); err != nil {
		a.fakeBcryptRounds()
		return err
	}

	incomingPassword, err := bcrypt.GenerateFromPassword([]byte(incoming), bcryptCostFactor)
	if err != nil {
		a.fakeBcryptRounds()
		return err // this path takes less time (no CompareHashAndPassword call)
	}
	return bcrypt.CompareHashAndPassword([]byte(storedPassword+storedSalt), incomingPassword)
}

func (a *auth) writePassword(userId string, pass string) error {
	salt := generateID()
	incomingPassword, err := bcrypt.GenerateFromPassword([]byte(pass+salt), bcryptCostFactor)
	if err != nil {
		return err
	}

	stmt, err := a.db.Prepare(`replace into user_passwords (user_id, password, salt) values (?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(userId, string(incomingPassword), salt)
	if err != nil {
		return err
	}
	a.log.Log("user", fmt.Sprintf("userId=%s updated password", userId))
	return nil
}

func hash(in string) (string, error) {
	ss := sha256.New()
	n, err := ss.Write([]byte(in))
	if err != nil {
		return "", err
	}
	if n == 0 {
		return "", err
	}
	return hex.EncodeToString(ss.Sum(nil)), nil
}
