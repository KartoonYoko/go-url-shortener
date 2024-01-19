package shortener

import "context"

type ShortenerRepo interface {
	SaveURL(ctx context.Context, url string) (string, error)
	GetURLByID(ctx context.Context, id string) (string, error)
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
