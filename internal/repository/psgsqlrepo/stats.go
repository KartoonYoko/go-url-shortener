package psgsqlrepo

import (
	"context"
	"fmt"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/stats"
)

// GetStats возвращает статистику
func (s *psgsqlRepo) GetStats(ctx context.Context) (*model.StatsResponse, error) {
	var count int
	var query string
	var err error
	response := new(model.StatsResponse)

	query = `SELECT COUNT(*) FROM shorten_url`
	err = s.conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("can not count shorten_url: %w", err)
	}
	response.URLs = count

	query = `SELECT COUNT(*) FROM users`
	err = s.conn.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("can not count users: %w", err)
	}
	response.Users = count

	return response, nil
}
