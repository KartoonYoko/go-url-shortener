package psgsqlrepo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type psgsqlRepo struct {
	conn *sqlx.DB
}

// NewPsgsqlRepo инициализирует хранилище для работы с БД
func NewPsgsqlRepo(ctx context.Context, db *sqlx.DB) (*psgsqlRepo, error) {
	repo := &psgsqlRepo{
		conn: db,
	}

	return repo, nil
}

// Close релизует Closer
func (s *psgsqlRepo) Close() error {
	s.conn.Close()
	return nil
}
