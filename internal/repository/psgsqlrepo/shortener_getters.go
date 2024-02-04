package psgsqlrepo

import (
	"context"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
)

func (s *psgsqlRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	row := s.conn.QueryRowContext(ctx, "SELECT url FROM shorten_url WHERE id=$1", id)
	var url string
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

// GetUserURLs вернёт все когда-либо сокращенные URL'ы пользователем
func (s *psgsqlRepo) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
	type GetModel struct {
		URLID string `db:"url_id"`
		URL   string `db:"url"`
	}
	models := []GetModel{}
	err := s.conn.Select(&models, `
	SELECT url_id, url FROM users_shorten_url 
	LEFT JOIN shorten_url ON shorten_url.id=users_shorten_url.url_id
	WHERE user_id=$1
	`, userID)

	if err != nil {
		return nil, err
	}

	response := make([]model.GetUserURLsItemResponse, 0, len(models))
	for _, v := range models {
		response = append(response, model.GetUserURLsItemResponse{
			ShortURL:    v.URLID,
			OriginalURL: v.URL,
		})
	}

	return response, nil
}
