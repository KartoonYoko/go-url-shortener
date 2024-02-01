package shortener

import (
	"context"
	"crypto/sha256"
	"math/rand"
	"time"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/google/uuid"
)

// данные url'а
type urlDataItem struct {
	url   string              // оригинальный URL
	users map[string]struct{} // пользователи, которые когда-либо формировали этот URL; 
}

// хранилище коротки адресов в памяти
type inMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - информация об URL'е
	storage map[string]urlDataItem
	r       *rand.Rand
}

func NewInMemoryRepo() *inMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]urlDataItem)
	return &inMemoryRepo{
		storage: s,
		r:       r,
	}
}

// сохранит url и вернёт его id'шник
func (s *inMemoryRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	h := sha256.New()
	hash, err := generateURLUniqueHash(h, url)
	if err != nil {
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

func (s *inMemoryRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	res, ok := s.storage[id]

	if !ok {
		return "", ErrNotFoundKey
	}

	return res.url, nil
}

func (s *inMemoryRepo) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
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

func (s *inMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *inMemoryRepo) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, len(request))
	for _, v := range request {
		hash, err := s.SaveURL(ctx, v.OriginalURL, userID)
		if err != nil {
			return nil, err
		}

		response = append(response, model.CreateShortenURLBatchItemResponse{
			CorrelationID: v.CorrelationID,
			ShortURL:      hash,
		})
	}

	return response, nil
}

func (s *inMemoryRepo) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

func (s *inMemoryRepo) Close() error {
	return nil
}
