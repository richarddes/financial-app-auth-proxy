package models

import (
	"auth-proxy/config"
	"context"
	"database/sql"
	"errors"
)

// Login returns the uid, saved password and the user's language, in that order, when
// the specified credentials match an entry. Otherwise it returns a config.ErrBadRequest.
func (db *DB) Login(ctx context.Context, body config.LoginReqBody) (uint64, string, string, error) {
	if body.Email == "" || body.Pass == "" {
		return 0, "", "", errors.New("Not all fields have been specified")
	}

	var (
		saveduid  uint64
		savedPass string
		lang      string
	)

	stmt := "SELECT id,pass,lang FROM users WHERE email=$1;"

	err := db.QueryRowContext(ctx, stmt, body.Email).Scan(&saveduid, &savedPass, &lang)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, "", "", config.ErrBadRequest
		}

		return 0, "", "", err
	}

	return saveduid, savedPass, lang, nil
}
