package psgsqlrepo

import (
	"context"
	"crypto/sha256"
	"errors"
	"strconv"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	repoCommon "github.com/KartoonYoko/go-url-shortener/internal/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// сохранит url и вернёт его id'шник
func (s *psgsqlRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	// сгенерируем уникальный ID для URL'a
	h := sha256.New()
	var err error
	hash, err := repoCommon.GenerateURLUniqueHash(h, url)
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

			err = repoCommon.NewURLAlreadyExistsError(hash, url)
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
	arrOfmapToInsert := []map[string]interface{}{}
	h := sha256.New()
	for _, url := range notExistsURLs {
		// если уже существует - добавлять не нужно
		if _, ok := existsURLs[url]; ok {
			continue
		}

		id, err := repoCommon.GenerateURLUniqueHash(h, url)
		if err != nil {
			return nil, err
		}

		arrOfmapToInsert = append(arrOfmapToInsert, map[string]interface{}{
			"id":  id,
			"url": url,
		})

		// запомним сгенерированный url для ответа и чтобы больше не генерировать ID
		existsURLs[url] = id
	}
	if len(arrOfmapToInsert) > 0 {
		_, err = s.conn.NamedExec(`INSERT INTO shorten_url (id, url) VALUES(:id, :url)`, arrOfmapToInsert)
		if err != nil {
			return nil, err
		}
	}

	// сохраним информацию о пользователе
	allURLsIDsMap := make(map[string]struct{})
	for _, urlHash := range existsURLs {
		allURLsIDsMap[urlHash] = struct{}{}
	}
	for _, mapItem := range arrOfmapToInsert {
		urlHashInterface, ok := mapItem["id"]
		if !ok {
			continue
		}
		urlHash, ok := urlHashInterface.(string)
		if !ok {
			continue
		}
		allURLsIDsMap[urlHash] = struct{}{}
	}
	allURLsIDsArr := make([]string, 0, len(allURLsIDsMap))
	for k := range allURLsIDsMap {
		allURLsIDsArr = append(allURLsIDsArr, k)
	}
	err = s.insertUserIDAndHashes(ctx, userID, allURLsIDsArr)
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
		UserID string `db:"user_id"`
		URLID  string `db:"url_id"`
	}
	hashesToInsert := make([]insertUserURLModel, 0, len(hashes))
	for _, urlID := range hashes {
		if _, ok := existsUserHashes[urlID]; ok {
			continue
		}

		hashesToInsert = append(hashesToInsert, insertUserURLModel{
			UserID: userID,
			URLID:  urlID,
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
	query, args, err := sqlx.In(`SELECT id, url FROM shorten_url WHERE url IN (?)`, requestURLs)
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
