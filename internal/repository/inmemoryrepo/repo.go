/*
Package inmemoryrepo релизация хранилища URL'ов в памяти
*/
package inmemoryrepo

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand"
	"time"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	modelStats "github.com/KartoonYoko/go-url-shortener/internal/model/stats"
	repoCommon "github.com/KartoonYoko/go-url-shortener/internal/repository"
	"github.com/google/uuid"
)

// данные url'а
type urlDataItem struct {
	url   string              // оригинальный URL
	users map[string]struct{} // пользователи, которые когда-либо формировали этот URL;
}

// InMemoryRepo хранилище коротких адресов в памяти
type InMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - информация об URL'е
	storage map[string]urlDataItem
	r       *rand.Rand
}

// NewInMemoryRepo инициализирует inmermory хранилище
func NewInMemoryRepo() *InMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]urlDataItem)
	return &InMemoryRepo{
		storage: s,
		r:       r,
	}
}

// UpdateURLsDeletedFlag пометит URL'ы удалёнными
func (s *InMemoryRepo) UpdateURLsDeletedFlag(ctx context.Context, userID string, modelsCh <-chan model.UpdateURLDeletedFlag) error {
	return errors.New("not implemented")
}

// SaveURL сохранит url и вернёт его id'шник
func (s *InMemoryRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	h := sha256.New()
	hash, err := repoCommon.GenerateURLUniqueHash(h, url)
	if err != nil {
		return "", err
	}

	// если уже существует
	if _, ok := s.storage[hash]; ok {
		err = repoCommon.NewURLAlreadyExistsError(hash, url)
		return "", err
	}

	data := urlDataItem{
		url:   url,
		users: map[string]struct{}{},
	}
	if _, ok := data.users[userID]; !ok {
		if userID != "" {
			data.users[userID] = struct{}{}
		}
	}

	s.storage[hash] = data
	return hash, nil
}

// GetURLByID вернёт URL по ID
func (s *InMemoryRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	res, ok := s.storage[id]

	if !ok {
		return "", repoCommon.ErrNotFoundKey
	}

	return res.url, nil
}

// GetUserURLs вернёт все URL'ы пользователя
func (s *InMemoryRepo) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
	response := make([]model.GetUserURLsItemResponse, 0)
	for urlID, data := range s.storage {
		if _, ok := data.users[userID]; !ok {
			continue
		}

		response = append(response, model.GetUserURLsItemResponse{
			OriginalURL: data.url,
			ShortURL:    urlID,
		})
	}

	return response, nil
}

// Ping реализует интерфейс Pinger
func (s *InMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

// SaveURLsBatch сохранит множество URL'ов пачкой
func (s *InMemoryRepo) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, 0, len(request))
	for _, v := range request {
		hash, err := s.SaveURL(ctx, v.OriginalURL, userID)
		if err != nil {
			var errAlreadyExists *repoCommon.URLAlreadyExistsError
			if errors.As(err, &errAlreadyExists) {
				response = append(response, model.CreateShortenURLBatchItemResponse{
					CorrelationID: v.CorrelationID,
					ShortURL:      errAlreadyExists.ID,
				})

				continue
			}
			return nil, err
		}

		response = append(response, model.CreateShortenURLBatchItemResponse{
			CorrelationID: v.CorrelationID,
			ShortURL:      hash,
		})
	}

	return response, nil
}

// GetNewUserID вернёт новый уникальны ID
func (s *InMemoryRepo) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

// Close релизует Closer
func (s *InMemoryRepo) Close() error {
	return nil
}

// Clear удалит все данные из хранилища
func (s *InMemoryRepo) Clear() error {
	for k := range s.storage {
		delete(s.storage, k)
	}

	return nil
}

// GetStats возвращает статистику
func (s *InMemoryRepo) GetStats(ctx context.Context) (*modelStats.StatsResponse, error) {
	users := make(map[string]struct{})
	for _, v := range s.storage {
		for k := range v.users {
			users[k] = struct{}{}
		}
	}

	response := new(modelStats.StatsResponse)
	response.URLs = len(s.storage)
	response.Users = len(users)

	return response, nil
}
