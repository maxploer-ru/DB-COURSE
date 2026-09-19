package repository

import (
	"errors"
	"strings"

	"ZVideo/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate key")
}

func mapUniqueConstraint(err error, constraints map[string]error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if mapped, ok := constraints[pgErr.ConstraintName]; ok {
			return mapped
		}
	}

	// Keep a fallback for drivers that wrap PostgreSQL errors as strings.
	message := strings.ToLower(err.Error())
	for constraint, mapped := range constraints {
		if strings.Contains(message, strings.ToLower(constraint)) {
			return mapped
		}
	}
	return nil
}

func mapUserConstraintError(err error) error {
	return mapUniqueConstraint(err, map[string]error{
		"users_username_key": domain.ErrUserNameAlreadyExists,
		"users_email_key":    domain.ErrUserEmailAlreadyExists,
	})
}
