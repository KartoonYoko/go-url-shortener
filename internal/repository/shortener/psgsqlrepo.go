package repository

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type psgsqlRepo struct {
	db *sql.DB
}

func NewPsgsqlRepo(db *sql.DB) (*psgsqlRepo, error) {
	return &psgsqlRepo{
		db: db,
	}, nil
}

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(url string) (string, error) {

	return "", nil
}

func (s *psgsqlRepo) GetURLByID(id string) (string, error) {
	return "", nil
}

func (s *psgsqlRepo) Close() error {
	s.db.Close()
	return nil
}

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
