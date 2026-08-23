package store

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func isUnique(err error) bool {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		return pg.Code == "23505"
	}
	return strings.Contains(err.Error(), "duplicate key")
}
