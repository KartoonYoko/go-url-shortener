package psgsqlrepo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// NewSQLxConnection создаёт экземпляр sqlx.DB и накатывает миграции
func NewSQLxConnection(ctx context.Context, databaseDsn string) (*sqlx.DB, error) {
	_, cancel := context.WithTimeout(ctx, time.Second*100)
	defer cancel()
	db, err := sqlx.Connect("pgx", databaseDsn)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
