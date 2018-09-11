// Copyright 2018 The ACH Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	// db is the connection point for making SQL calls to our sqlite database.
	// You use it like any other database/sql driver. Part of shutdown is to close
	// out the file/session.
	//
	// https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
	// https://astaxie.gitbooks.io/build-web-application-with-golang/en/05.3.html
	db         *sql.DB
	sqlitePath string

	migrations = []string{
		// Initial user setup
		`create table if not exists users(user_id primary key, email, clean_email, password, salt, created_at timestamp);`,
		`create table if not exists user_details(user_id primary key, first_name, last_name, phone, company_url);`,
		`create table if not exists user_approval_codes (user_id, code, valid_until);`,
	}
)

// TODO(adam): prometheus metrics
// $ go doc database/sql dbstats
// type DBStats struct {
// 	MaxOpenConnections int // Maximum number of open connections to the database.
// 	OpenConnections int // The number of established connections both in use and idle.
// 	InUse           int // The number of connections currently in use.
// 	Idle            int // The number of idle connections.
// }

func init() {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" || strings.Contains(path, "..") {
		// set default if empty or trying to escape
		// don't filepath.ABS to avoid full-fs reads
		path = "auth.db"
	}

	d, err := sql.Open("sqlite3", path)
	if err != nil {
		err = fmt.Errorf("problem opening sqlite3 file: %v", err)
		logger.Log("sqlite", err)
		panic(err.Error())
	}
	db = d
	sqlitePath = path
}

func migrate() error {
	logger.Log("sqlite", fmt.Sprintf("migrating %s", sqlitePath))
	for i := range migrations {
		row := migrations[i]
		res, err := db.Exec(row)
		if err != nil {
			return fmt.Errorf("migration #%d [%s...] had problem: %v", i, row[:40], err)
		}
		n, err := res.RowsAffected()
		if err == nil {
			logger.Log("sqlite", fmt.Sprintf("migration #%d [%s...] changed %d rows", i, row[:40], n))
		}
	}
	logger.Log("sqlite", "finished migrations")
	return nil
}
