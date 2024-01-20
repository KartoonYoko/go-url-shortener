package shortener

import (
	"context"
	"crypto/sha256"
	"errors"
	"strconv"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"github.com/KartoonYoko/go-url-shortener/internal/model"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
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

	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
		id VARCHAR PRIMARY KEY,
		url VARCHAR
	)`)
	tx.ExecContext(ctx, "CREATE UNIQUE INDEX url_idx ON shorten_url (url)")

	return tx.Commit()
}

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(ctx context.Context, url string) (string, error) {
	hash := randStringRunes(5)
	_, err := s.conn.ExecContext(ctx, "INSERT INTO shorten_url (url, id) VALUES($1, $2)", url, hash)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func (s *psgsqlRepo) SaveURLsBatch(ctx context.Context,
	batch []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error) {
	// проверим какие URL'ы существуют;
	// все существующие URL'ы добавим в словарь,
	// где ключ - URL, значение - ID;
	existsURLs, err := s.getMapedExistsURLs(ctx, batch)
	if err != nil {
		return nil, err
	}

	// запомним несуществующие URL'ы
	notExistsURLs := make([]string, len(batch))
	for _, v := range batch {
		if _, ok := existsURLs[v.OriginalURL]; ok {
			continue
		}

		notExistsURLs = append(notExistsURLs, v.OriginalURL)
	}

	// добавим в БД несуществующие
	mapToInsert := []map[string]interface{}{}
	h := sha256.New()
	for _, url := range notExistsURLs {
		// если уже существует - добавлять не нужно
		if _, ok := existsURLs[url]; ok {
			continue
		}
		h.Reset()
		// сгенерируем уникальный ID для URL'a
		_, err := h.Write([]byte(url))
		if err != nil {
			return nil, err
		}
		id := string(h.Sum(nil))

		mapToInsert = append(mapToInsert, map[string]interface{}{
			"id":  id,
			"url": url,
		})

		// запомним сгенерированный url для ответа и чтобы больше не генерировать ID
		existsURLs[url] = id
	}
	_, err = s.conn.NamedExec(`INSERT INTO shorten_url (id, url) VALUES(:id, :url)`, mapToInsert)
	if err != nil {
		return nil, err
	}

	// соберём ответ
	response := make([]model.CreateShortenURLBatchItemResponse, len(existsURLs))
	for url, id := range existsURLs {
		// найдём все CorrelationID с данным URL'ом
		сorrelationIDs := []string{}
		for _, requestItem := range batch {
			if requestItem.OriginalURL == url {
				сorrelationIDs = append(сorrelationIDs, requestItem.CorrelationID)
			}
		}

		// соберём ответы для каждого сorrelationID
		for _, сorrelationID := range сorrelationIDs {
			response = append(response, model.CreateShortenURLBatchItemResponse{
				ShortURL:      id,
				CorrelationID: сorrelationID,
			})
		}
	}

	// если количество ответов не совпало с количеством запросов - выдаём ошибку
	if len(batch) != len(response) {
		message := "batch request to insert url's returns wrong responses count"
		responseCountZapStr := zap.String("response count", strconv.Itoa(len(response)))
		requestCountZapStr := zap.String("request count", strconv.Itoa(len(batch)))
		logger.Log.Error(
			message,
			requestCountZapStr,
			responseCountZapStr)

		return nil, errors.New(message)
	}

	return response, nil
}

func (s *psgsqlRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	row := s.conn.QueryRowContext(ctx, "SELECT url FROM shorten_url WHERE id=$1", id)
	var url string
	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (s *psgsqlRepo) Close() error {
	s.conn.Close()
	return nil
}

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

// getMapedExistsURLs вернёт существующие URL'ы в виде словаря,
// где ключ - URL, значение - ID этого URL'а
func (s *psgsqlRepo) getMapedExistsURLs(ctx context.Context,
	batch []model.CreateShortenURLBatchItemRequest) (map[string]string, error) {
	// подготовим запрос для нахождения всех URL'ов
	requestURLs := make([]string, len(batch))
	for _, v := range batch {
		requestURLs = append(requestURLs, v.OriginalURL)
	}
	query, args, err := sqlx.In(`SELECT * FROM shorten_url WHERE url IN (?)`, requestURLs)
	if err != nil {
		return nil, err
	}
	query = s.conn.Rebind(query)
	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// все существующие URL'ы добавим в словарь,
	// где ключ - URL, значение - ID
	existsURLs := map[string]string{}
	for rows.Next() {
		var id, url string
		err = rows.Scan(&id, &url)
		if err != nil {
			return nil, err
		}

		existsURLs[id] = url
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return existsURLs, nil
}
