package grpcserver

import (
	"context"

	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	modelStats "github.com/KartoonYoko/go-url-shortener/internal/model/stats"
)

type UseCaseShortener interface {
	GetURLByID(ctx context.Context, urlID string) (string, error)
	SaveURL(ctx context.Context, url string, userID string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error)
	DeleteURLs(ctx context.Context, userID string, urlsIDs []string) error
}

type UseCasePinger interface {
	Ping(ctx context.Context) error
}

type useCaseAuther interface {
	GetNewUserID(ctx context.Context) (string, error)
}

type UseCaseStats interface {
	GetStats(ctx context.Context) (*modelStats.StatsResponse, error)
}
