package storage

import (
	"errors"
	"strings"

	"github.com/lib/pq"
)

// isUniqueViolation reports whether err is a unique-constraint violation, handling
// both Postgres (pq error code 23505) and SQLite (the driver surfaces a "UNIQUE
// constraint failed" message). Stores translate this into their feature's typed
// "already exists" / "duplicate" error.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
