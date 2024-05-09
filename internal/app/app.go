/*
Package app предоставляет методы запуска приложения
*/
package app

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	fileRepo "github.com/KartoonYoko/go-url-shortener/internal/repository/filerepo"
	inmrRepo "github.com/KartoonYoko/go-url-shortener/internal/repository/inmemoryrepo"
	pgsqlRepo "github.com/KartoonYoko/go-url-shortener/internal/repository/psgsqlrepo"
	usecaseAuth "github.com/KartoonYoko/go-url-shortener/internal/usecase/auth"
	usecasePinger "github.com/KartoonYoko/go-url-shortener/internal/usecase/ping"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"
	usecaseStats "github.com/KartoonYoko/go-url-shortener/internal/usecase/stats"
	"go.uber.org/zap"
)

// ShortenerRepoCloser интерфейс, объединяющий в себе все необходимые репозитории
type ShortenerRepoCloser interface {
	usecaseShortener.ShortenerRepo
	usecasePinger.PingRepo
	usecaseAuth.AuthRepo
	usecaseStats.StatsRepo
	io.Closer
}

// Run запускает приложение
func Run() {
	ctx := context.TODO()

	// логгер
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(fmt.Errorf("logger init error: %w", err))
	}
	defer logger.Log.Sync()

	// конфигурация
	conf, err := config.New()
	if err != nil {
		logger.Log.Error("config init error: ", zap.Error(err))
		return
	}

	// репозитории
	repo, err := initRepo(ctx, *conf)
	if err != nil {
		logger.Log.Error("repo init error: %s", zap.Error(err))
		return
	}
	defer repo.Close()

	// usecase'ы
	serviceShortener := usecaseShortener.New(repo, conf.BaseURLAddress)
	servicePinger := usecasePinger.NewPingUseCase(repo)
	serviceAuth := usecaseAuth.NewAuthUseCase(repo)
	serviceStats := usecaseStats.New(repo)

	// контроллеры
	shortenerController := http.NewShortenerController(
		serviceShortener,
		servicePinger,
		serviceAuth,
		serviceStats,
		conf)

	err = shortenerController.Serve(ctx)
	if err != nil {
		logger.Log.Error("serve error: %s", zap.Error(err))
		return
	}
}

func initRepo(ctx context.Context, conf config.Config) (ShortenerRepoCloser, error) {
	if conf.DatabaseDsn != "" {
		db, err := pgsqlRepo.NewSQLxConnection(ctx, conf.DatabaseDsn)
		if err != nil {
			return nil, err
		}

		repo, err := pgsqlRepo.NewPsgsqlRepo(ctx, db)
		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	if conf.FileStoragePath != "" {
		fileRepo, err := fileRepo.NewFileRepo(conf.FileStoragePath)
		if err != nil {
			return nil, err
		}

		return fileRepo, nil
	}

	return inmrRepo.NewInMemoryRepo(), nil
}
