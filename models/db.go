// Package models implements functions manipulating the database.
package models

import (
	"database/sql"
	"time"
)

// DB represents a database connection.
type DB struct {
	*sql.DB
}

var retryCount = 1

// New returns a new instance of the DB struct. It tries to connect to
// the DB 5 times by waiting 10s after each failed attempt.
// If it fails 5 times to connect it returns an error.
func New(dbPath string) (*DB, error) {
	db, err := sql.Open("postgres", dbPath)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 10)
	if err = db.Ping(); err != nil {
		if retryCount <= 5 {
			retryCount++
			New(dbPath)
		}

		return nil, err
	}

	return &DB{db}, nil
}
