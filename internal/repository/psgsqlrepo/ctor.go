package psgsqlrepo

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type psgsqlRepo struct {
	conn *sqlx.DB
}

func NewPsgsqlRepo(ctx context.Context, db *sqlx.DB) (*psgsqlRepo, error) {
	repo := &psgsqlRepo{
		conn: db,
	}
	err := repo.createSchema(ctx)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (s *psgsqlRepo) createSchema(ctx context.Context) (err error) {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer tx.Rollback()

	// таблица URL'ов
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
		id VARCHAR PRIMARY KEY,
		url VARCHAR,
		deleted_flag boolean DEFAULT false,
	)`)
	tx.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS url_idx ON shorten_url (url)")

	// таблица пользователей
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (
		id VARCHAR PRIMARY KEY
	)`)

	// таблица пользователей/URL'ов
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users_shorten_url (
		user_id VARCHAR,
		url_id VARCHAR,

		PRIMARY KEY(user_id, url_id),

		CONSTRAINT fk_user_id
		FOREIGN KEY (user_id) 
		REFERENCES users (id),

		CONSTRAINT fk_url_id
		FOREIGN KEY (url_id) 
		REFERENCES shorten_url (id)
	)`)

	return tx.Commit()
}

func (s *psgsqlRepo) Close() error {
	s.conn.Close()
	return nil
}
