package app

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	repository "github.com/KartoonYoko/go-url-shortener/internal/repository/shortener"
	usecasePinger "github.com/KartoonYoko/go-url-shortener/internal/usecase/ping"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"
	usecaseAuth "github.com/KartoonYoko/go-url-shortener/internal/usecase/auth"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type ShortenerRepoCloser interface {
	usecaseShortener.ShortenerRepo
	usecasePinger.PingRepo
	usecaseAuth.AuthRepo
	io.Closer
}

func Run() {
	ctx := context.TODO()

	// логгер
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(fmt.Errorf("logger init error: %w", err))
	}
	defer logger.Log.Sync()
	conf := config.New()

	// репозитории
	repo, err := initRepo(ctx, *conf)
	if err != nil {
		log.Fatal(fmt.Errorf("repository init error: %w", err))
	}
	defer repo.Close()

	// usecase'ы
	serviceShortener := usecaseShortener.New(repo, conf.BaseURLAddress)
	servicePinger := usecasePinger.NewPingUseCase(repo)
	serviceAuth := usecaseAuth.NewAuthUseCase(repo)

	// контроллеры
	shortenerController := http.NewShortenerController(serviceShortener, servicePinger, serviceAuth, conf)

	shortenerController.Serve()
}

func initRepo(ctx context.Context, conf config.Config) (ShortenerRepoCloser, error) {
	if conf.DatabaseDsn != "" {
		db, err := sqlx.Connect("pgx", conf.DatabaseDsn)
		if err != nil {
			return nil, err
		}

		repo, err := repository.NewPsgsqlRepo(ctx, db)
		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	if conf.FileStoragePath != "" {
		fileRepo, err := repository.NewFileRepo(conf.FileStoragePath)
		if err != nil {
			return nil, err
		}

		return fileRepo, nil
	}

	return repository.NewInMemoryRepo(), nil
}
