package shortener

import (
	"context"
	"errors"
	"fmt"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	repository "github.com/KartoonYoko/go-url-shortener/internal/repository/shortener"
	"go.uber.org/zap"
)

type ShortenerRepo interface {
	SaveURL(ctx context.Context, url string, userID string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error)
	GetURLByID(ctx context.Context, id string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error)
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
func (s *shortenerUsecase) SaveURL(ctx context.Context, hash string, userID string) (string, error) {
	hash, err := s.repository.SaveURL(ctx, hash, userID)
	if err != nil {
		var repoErrURLAlreadyExists *repository.URLAlreadyExistsError
		if errors.As(err, &repoErrURLAlreadyExists) {
			return "", NewURLAlreadyExistsError(s.getShorURL(repoErrURLAlreadyExists.ID), repoErrURLAlreadyExists.URL, err)
		}
		logger.Log.Error("save url error", zap.Error(err))
		return "", err
	}

	return s.getShorURL(hash), nil
}

func (s *shortenerUsecase) GetURLByID(ctx context.Context, id string) (string, error) {
	url, err := s.repository.GetURLByID(ctx, id)
	if err != nil {
		logger.Log.Error("get url error", zap.Error(err))
		return "", err
	}
	return url, nil
}

func (s *shortenerUsecase) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
	res, err := s.repository.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range res {
		res[i].ShortURL = s.getShorURL(res[i].ShortURL)
	}

	return res, nil
}

func (s *shortenerUsecase) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	response, err := s.repository.SaveURLsBatch(ctx, request, userID)
	if err != nil {
		logger.Log.Error("save urls batch error", zap.Error(err))
		return nil, err
	}

	for i := range response {
		response[i].ShortURL = s.getShorURL(response[i].ShortURL)
	}
	return response, nil
}

func (s *shortenerUsecase) getShorURL(hash string) string {
	return fmt.Sprintf("%s/%s", s.baseURLAddress, hash)
}
