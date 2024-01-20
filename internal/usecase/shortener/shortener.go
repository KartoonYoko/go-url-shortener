package shortener

import (
	"context"
	"fmt"

	"github.com/KartoonYoko/go-url-shortener/internal/model"
)

type ShortenerRepo interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURLByID(ctx context.Context, id string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error)
}

type shortenerUsecase struct {
	repository     ShortenerRepo
	baseURLAddress string // Базовый адрес результирующего сокращенного URL
}

func New(repo ShortenerRepo, baseURLAddress string) *shortenerUsecase {
	return &shortenerUsecase{
		repository:     repo,
		baseURLAddress: baseURLAddress,
	}
}

// сохранит url и вернёт его id'шник
func (s *shortenerUsecase) SaveURL(ctx context.Context, hash string) (string, error) {
	hash, err := s.repository.SaveURL(ctx, hash)
	if err != nil {
		return "", err
	}

	return s.getShorURL(hash), nil
}

func (s *shortenerUsecase) GetURLByID(ctx context.Context, id string) (string, error) {
	hash, err := s.repository.GetURLByID(ctx, id)
	if err != nil {
		return "", err
	}
	return s.getShorURL(hash), nil
}

func (s *shortenerUsecase) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error) {
	response, err := s.repository.SaveURLsBatch(ctx, request)
	if err != nil {
		return nil, err
	}

	for _, v := range response {
		v.ShortURL = s.getShorURL(v.ShortURL)
	}
	return response, nil
}

func (s *shortenerUsecase) getShorURL(hash string) string {
	return fmt.Sprintf("%s/%s", s.baseURLAddress, hash)
}
