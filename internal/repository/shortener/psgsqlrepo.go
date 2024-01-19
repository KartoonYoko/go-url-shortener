package shortener

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type psgsqlRepo struct {
	db *pgxpool.Pool
}

func NewPsgsqlRepo(ctx context.Context, db *pgxpool.Pool) (*psgsqlRepo, error) {
	repo := &psgsqlRepo{
		db: db,
	}
	err := repo.createSchema(ctx, db)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *psgsqlRepo) createSchema(ctx context.Context, db *pgxpool.Pool) error {
	schema := `CREATE TABLE IF NOT EXISTS shorten_url (
		url VARCHAR,
		id VARCHAR
	)`

	_, err := db.Exec(ctx, schema)

	return err
}

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(ctx context.Context, url string) (string, error) {
	hash := randStringRunes(5)
	_, err := s.db.Exec(ctx, "INSERT INTO shorten_url (url, id) VALUES($1, $2)", url, hash)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func (s *psgsqlRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	row := s.db.QueryRow(ctx, "SELECT url FROM shorten_url WHERE id=$1", id)
	var url string
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *psgsqlRepo) Close() error {
	s.db.Close()
	return nil
}

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
