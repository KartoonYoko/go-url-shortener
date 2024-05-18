package stats

import (
	"context"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/stats"
	"go.uber.org/zap"
)

// StatsRepo интерфейс хранилища
type StatsRepo interface {
	GetStats(ctx context.Context) (*model.StatsResponse, error)
}

type statsUsecase struct {
	repository StatsRepo
}

// New инициализирует shortenerUsecase
func New(repo StatsRepo) *statsUsecase {
	uc := new(statsUsecase)
	uc.repository = repo
	return uc
}

// GetStats возвращает статистику
func (s *statsUsecase) GetStats(ctx context.Context) (*model.StatsResponse, error) {
	res, err := s.repository.GetStats(ctx)
	if err != nil {
		logger.Log.Error("can not get stats",
			zap.String("package", "stats"),
			zap.String("func", "GetStats"),
			zap.Error(err))
		return nil, err
	}

	return res, nil
}
