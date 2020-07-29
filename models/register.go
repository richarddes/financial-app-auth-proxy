package models

import (
	"auth-proxy/config"
	"context"
	"errors"

	"github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

// Register adds a user to the db. It defaults the language to "en" (English) and the cash to 1000$.
// If the user already exists in the db it returns a config.ErrBadRequest.
func (db *DB) Register(ctx context.Context, body config.RegistrationReqBody) error {
	if body.Email == "" || body.Pass == "" || body.LastName == "" {
		return errors.New("Not all required fields (Email, Pass, LastName) have been specified")
	}

	stmt := "INSERT INTO users VALUES (DEFAULT,$1,$2,$3,$4,$5,$6,$7::stock[]);"

	passHash, err := bcrypt.GenerateFromPassword([]byte(body.Pass), bcrypt.DefaultCost)
	if err != nil {
		return nil
	}

	// Default values for every new user
	lang := "en"
	cash := 1000

	_, err = db.ExecContext(ctx, stmt, body.Email, string(passHash), body.LastName, body.FirstName, lang, cash, "{}")
	if err != nil {
		// If the specified email already exists, the db signals that it's not a server error but a user error.
		// In theory it also returns a unique_violation when the password already exists but that's unlikely to happen
		// so we use it to check if the email already exists.
		if err, ok := err.(*pq.Error); ok && err.Code.Name() == "unique_violation" {
			return config.ErrBadRequest
		}

		return err
	}

	return nil
}
