package app

import (
	"context"
	"io"
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	repository "github.com/KartoonYoko/go-url-shortener/internal/repository/shortener"
	usecasePinger "github.com/KartoonYoko/go-url-shortener/internal/usecase/ping"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ShortenerRepoCloser interface {
	usecaseShortener.ShortenerRepo
	usecasePinger.PingRepo
	io.Closer
}

func Run() {
	ctx := context.TODO()

	// логгер
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()
	conf := config.New()

	// репозитории
	repo, err := initRepo(ctx, *conf)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()

	// usecase'ы
	serviceShortener := usecaseShortener.New(repo)
	servicePinger := usecasePinger.NewPingUseCase(repo)

	// контроллеры
	shortenerController := http.NewShortenerController(serviceShortener, servicePinger, conf)

	shortenerController.Serve()
}

func initRepo(ctx context.Context, conf config.Config) (ShortenerRepoCloser, error) {
	if conf.DatabaseDsn != "" {
		db, err := pgxpool.New(ctx, conf.DatabaseDsn)
		if err != nil {
			return nil, err
		}

		repo, err := repository.NewPsgsqlRepo(ctx, db)
		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	fileRepo, err := repository.NewFileRepo(conf.FileStoragePath)
	if err == nil {
		return fileRepo, nil
	}

	inMemoryRepo := repository.NewInMemoryRepo()
	
	return inMemoryRepo, nil
}
