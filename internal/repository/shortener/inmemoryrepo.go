package shortener

import (
	"context"
	"math/rand"
	"time"

	"github.com/KartoonYoko/go-url-shortener/internal/model"
)

// хранилище коротки адресов в памяти
type inMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - url
	storage map[string]string
	r       *rand.Rand
}

func NewInMemoryRepo() *inMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]string)
	return &inMemoryRepo{
		storage: s,
		r:       r,
	}
}

// сохранит url и вернёт его id'шник
func (s *inMemoryRepo) SaveURL(ctx context.Context, url string) (string, error) {
	hash := randStringRunes(5)
	s.storage[hash] = url
	return hash, nil
}

func (s *inMemoryRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, ErrNotFoundKey
	}

	return res, nil
}

func (s *inMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *inMemoryRepo) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, len(request))
	for _, v := range request {
		hash, err := s.SaveURL(ctx, v.OriginalURL)
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

func (s *inMemoryRepo) Close() error {
	return nil
}
