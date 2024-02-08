package inmemoryrepo

import (
	"context"
	"crypto/sha256"
	"errors"
	"math/rand"
	"time"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	repoCommon "github.com/KartoonYoko/go-url-shortener/internal/repository"
	"github.com/google/uuid"
)

// данные url'а
type urlDataItem struct {
	url   string              // оригинальный URL
	users map[string]struct{} // пользователи, которые когда-либо формировали этот URL;
}

// хранилище коротки адресов в памяти
type InMemoryRepo struct {
	// хранилище адресов и их id'шников; ключ - id, значение - информация об URL'е
	storage map[string]urlDataItem
	r       *rand.Rand
}

func NewInMemoryRepo() *InMemoryRepo {
	r := rand.New(rand.NewSource(time.Now().UnixMilli()))
	s := make(map[string]urlDataItem)
	return &InMemoryRepo{
		storage: s,
		r:       r,
	}
}

func (s *InMemoryRepo) UpdateURLsDeletedFlag(ctx context.Context, userID string, modelsCh <-chan model.UpdateURLDeletedFlag) error {
	return errors.New("not implemented")
}

// сохранит url и вернёт его id'шник
func (s *InMemoryRepo) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	h := sha256.New()
	hash, err := repoCommon.GenerateURLUniqueHash(h, url)
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

func (s *InMemoryRepo) GetURLByID(ctx context.Context, id string) (string, error) {
	res, ok := s.storage[id]

	if !ok {
		return "", repoCommon.ErrNotFoundKey
	}

	return res.url, nil
}

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

func (s *InMemoryRepo) Ping(ctx context.Context) error {
	return nil
}

func (s *InMemoryRepo) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	response := make([]model.CreateShortenURLBatchItemResponse, 0, len(request))
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

func (s *InMemoryRepo) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

func (s *InMemoryRepo) Close() error {
	return nil
}
