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
		short_index VARCHAR
	)`

	_, err := db.Exec(ctx, schema)

	return err
}

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(ctx context.Context, url string) (string, error) {
	s.db.Exec(ctx, "INSERT INTO shorten_url (url, short_index) VALUES($1, $2)")
	return "", nil
}

func (s *psgsqlRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (s *psgsqlRepo) Close() error {
	s.db.Close()
	return nil
}

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}
