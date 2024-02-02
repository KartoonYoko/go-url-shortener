package shortener

import (
	"context"
	"crypto/sha256"
	"errors"
	"strconv"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

	// таблица URL'ов
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS shorten_url (
		id VARCHAR PRIMARY KEY,
		url VARCHAR
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

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	// сгенерируем уникальный ID для URL'a
	h := sha256.New()
	var err error
	hash, err := generateURLUniqueHash(h, url)
	if err != nil {
		return "", err
	}

	_, err = s.conn.ExecContext(ctx, "INSERT INTO shorten_url (url, id) VALUES($1, $2)", url, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		// если вставка не удалась по причине, что уже существует такой URL в БД,
		// то делаем ещё один запрос для определения существующего ID
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			row := s.conn.QueryRowContext(ctx, "SELECT id FROM shorten_url WHERE url=$1", url)
			err = row.Err()
			if err != nil {
				return "", err
			}
			err = row.Scan(&hash)
			if err != nil {
				return "", err
			}

			err = s.insertUserIDAndHash(ctx, userID, hash)
			if err != nil {
				return "", err
			}

			err = NewURLAlreadyExistsError(hash, url)
			return hash, err
		}

		return "", err
	}

	err = s.insertUserIDAndHash(ctx, userID, hash)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// SaveURLsBatch выполняет множественную вставку
func (s *psgsqlRepo) SaveURLsBatch(ctx context.Context,
	batch []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	// проверим какие URL'ы существуют;
	// все существующие URL'ы добавим в словарь,
	// где ключ - URL, значение - ID;
	existsURLs, err := s.getMapedExistsURLs(ctx, batch)
	if err != nil {
		return nil, err
	}

	// запомним несуществующие URL'ы
	notExistsURLs := make([]string, 0, len(batch))
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

		id, err := generateURLUniqueHash(h, url)
		if err != nil {
			return nil, err
		}

		mapToInsert = append(mapToInsert, map[string]interface{}{
			"id":  id,
			"url": url,
		})

		// запомним сгенерированный url для ответа и чтобы больше не генерировать ID
		existsURLs[url] = id
	}
	if len(mapToInsert) > 0 {
		_, err = s.conn.NamedExec(`INSERT INTO shorten_url (id, url) VALUES(:id, :url)`, mapToInsert)
		if err != nil {
			return nil, err
		}
	}

	// сохраним информацию о пользователе
	err = s.insertUserIDAndHashes(ctx, userID, notExistsURLs)
	if err != nil {
		return nil, err
	}

	// соберём ответ
	response := make([]model.CreateShortenURLBatchItemResponse, 0, len(existsURLs))
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

func (s *psgsqlRepo) Close() error {
	s.conn.Close()
	return nil
}

func (s *psgsqlRepo) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *psgsqlRepo) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	_, err := s.conn.ExecContext(ctx, "INSERT INTO users (id) VALUES($1)", id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			return id.String(), nil
		}

		return "", err
	}
	return id.String(), nil
}

// insertUserIDAndHash вставляет запись о пользователе и URL'е в таблицу, если записи нет
func (s *psgsqlRepo) insertUserIDAndHash(ctx context.Context, userID string, hash string) error {
	_, err := s.conn.ExecContext(ctx, "INSERT INTO users_shorten_url (user_id, url_id) VALUES($1, $2)", userID, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
			// если уже существует, то добавлять не нужно
			return nil
		}

		return err
	}

	return nil
}

// insertUserIDAndHashes вставляет записи о пользователе и URL'ах в таблицу, если записей нет
func (s *psgsqlRepo) insertUserIDAndHashes(ctx context.Context, userID string, hashes []string) error {
	// получение только тех записей, которые ещё не существуют в БД
	rows, err := s.conn.QueryContext(ctx, `SELECT url_id FROM users_shorten_url WHERE user_id=$1`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	existsUserHashes := make(map[string]struct{})
	for rows.Next() {
		var urlID string
		err := rows.Scan(&urlID)
		if err != nil {
			return err
		}
		existsUserHashes[urlID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	type insertUserURLModel struct {
		userID string `db:"url_id"`
		urlID  string `db:"url_id"`
	}
	hashesToInsert := make([]insertUserURLModel, 0, len(hashes))
	for _, urlID := range hashes {
		if _, ok := existsUserHashes[urlID]; ok {
			continue
		}

		hashesToInsert = append(hashesToInsert, insertUserURLModel{
			userID: userID,
			urlID:  urlID,
		})
	}
	if len(hashesToInsert) == 0 {
		return nil
	}

	_, err = s.conn.NamedExec(`INSERT INTO users_shorten_url (user_id, url_id) 
		VALUES (:user_id, :url_id)`, hashesToInsert)

	if err != nil {
		return err
	}

	return nil
}

// getMapedExistsURLs вернёт существующие URL'ы в виде словаря,
// где ключ - URL, значение - ID этого URL'а
func (s *psgsqlRepo) getMapedExistsURLs(ctx context.Context,
	batch []model.CreateShortenURLBatchItemRequest) (map[string]string, error) {
	// подготовим запрос для нахождения всех URL'ов
	requestURLs := make([]string, 0, len(batch))
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

		existsURLs[url] = id
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return existsURLs, nil
}
