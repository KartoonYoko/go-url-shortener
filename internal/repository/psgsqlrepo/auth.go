package psgsqlrepo

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// GetNewUserID создаст нового пользователя ивернёт его ID
func (s *psgsqlRepo) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	_, err := s.conn.ExecContext(ctx, "INSERT INTO users (id) VALUES($1)", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return id.String(), nil
		}

		return "", err
	}
	return id.String(), nil
}
