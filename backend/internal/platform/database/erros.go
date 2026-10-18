package database

import (
	"errors"
	"strings"
)

// IsUniqueViolation reports whether err comes from a unique constraint (or a
// primary key) violation. It is how repositories translate a database race
// (comentário removido: texto original indisponível)
//
// The check matches both pgx and lib/pq error strings and PostgreSQL's
// SQLSTATE 23505, which is what the migrations rely on.
func EhViolacaoDeUnicidade(err error) bool {
	if err == nil {
		return false
	}

	// pgx v5 exposes the SQLSTATE through a typed error.
	var erroDoPG interface{ SQLState() string }
	if errors.As(err, &erroDoPG) {
		if erroDoPG.SQLState() == "23505" {
			return true
		}
	}

	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "unique constraint")
}

// IsForeignKeyViolation reports whether err comes from a broken foreign key
// (SQLSTATE 23503). Used to answer 409 instead of 500 when a delete would
// orphan children.
func EhViolacaoDeChaveEstrangeira(err error) bool {
	if err == nil {
		return false
	}
	var erroDoPG interface{ SQLState() string }
	if errors.As(err, &erroDoPG) {
		if erroDoPG.SQLState() == "23503" {
			return true
		}
	}
	msg := err.Error()
	return strings.Contains(msg, "SQLSTATE 23503") ||
		strings.Contains(msg, "foreign key constraint")
}

// IsCheckViolation reports whether err comes from a CHECK constraint
// (SQLSTATE 23514). The database-level role guards raise plain exceptions
// instead, so this covers only the data-integrity checks.
func EhViolacaoDeRestricao(err error) bool {
	if err == nil {
		return false
	}
	var erroDoPG interface{ SQLState() string }
	if errors.As(err, &erroDoPG) {
		if erroDoPG.SQLState() == "23514" {
			return true
		}
	}
	return strings.Contains(err.Error(), "SQLSTATE 23514")
}
