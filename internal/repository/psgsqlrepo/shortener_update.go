package psgsqlrepo

import (
	"context"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/jmoiron/sqlx"
)

func (s *psgsqlRepo) UpdateURLsDeletedFlag(ctx context.Context, userID string, modelsCh <-chan model.UpdateURLDeletedFlag) error {
	type argType struct {
		UserID string   `db:"user_id"`
		URLIDs []string `db:"url_ids"`
	}
	arg := argType{
		UserID: userID,
		URLIDs: make([]string, 0),
	}
	for model := range modelsCh {
		arg.URLIDs = append(arg.URLIDs, model.URLID)
	}

	query := `
	UPDATE shorten_url AS su
	SET deleted_flag = true
		FROM users_shorten_url AS usu
	WHERE usu.user_id=:user_id AND
	su.id IN (:url_ids)`

	query, args, err := sqlx.Named(query, arg)
	if err != nil {
		return err
	}
	query, args, err = sqlx.In(query, args...)
	if err != nil {
		return err
	}

	query = s.conn.Rebind(query)
	_, err = s.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
