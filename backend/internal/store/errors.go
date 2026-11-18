package store

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err is a natural-key duplicate.
//
// This is deliberately the only place in the codebase that inspects
// driver error text: "UNIQUE constraint failed" is modernc.org/sqlite's
// wording, and PostgreSQL (ADR-0001 exit path) says "duplicate key value
// violates unique constraint". Isolating it here means the swap touches
// exactly one function.
func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code == "23505"
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "duplicate key value violates unique constraint")
}
