package models

import (
	"context"
	"database/sql"
)

// UserCount returns the amount of times a user-id is found in a dataabase. In case of a table schema which hasn't marked
// the user-id field as a unique, this function can be used to identify if a user-id exists multiple times.
func UserCount(ctx context.Context, uid uint64, db *sql.DB) (int, error) {
	var n int

	query := `SELECT COUNT(id)
						FROM users
						WHERE	id=$1;`

	err := db.QueryRowContext(ctx, query, uid).Scan(&n)
	if err != nil {
		return -1, err
	}

	return n, nil
}
