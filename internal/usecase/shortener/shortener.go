package shortener

import (
	"context"

	"github.com/KartoonYoko/go-url-shortener/internal/model"
)

type ShortenerRepo interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURLByID(ctx context.Context, id string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error)
}

type shortenerUsecase struct {
	repository ShortenerRepo
}

func New(repo ShortenerRepo) *shortenerUsecase {
	return &shortenerUsecase{
		repository: repo,
	}
}

// сохранит url и вернёт его id'шник
func (s *shortenerUsecase) SaveURL(ctx context.Context, url string) (string, error) {
	return s.repository.SaveURL(ctx, url)
}

func (s *shortenerUsecase) GetURLByID(ctx context.Context, id string) (string, error) {
	return s.repository.GetURLByID(ctx, id)
}

func (s *shortenerUsecase) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest) ([]model.CreateShortenURLBatchItemResponse, error) {
	return s.repository.SaveURLsBatch(ctx, request)
}
