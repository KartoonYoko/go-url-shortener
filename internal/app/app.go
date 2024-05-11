/*
Package app предоставляет методы запуска приложения
*/
package app

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver"
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

// shortenerRepoCloser интерфейс, объединяющий в себе все необходимые репозитории
type shortenerRepoCloser interface {
	usecaseShortener.ShortenerRepo
	usecasePinger.PingRepo
	usecaseAuth.AuthRepo
	usecaseStats.StatsRepo
	io.Closer
}

type serverHandler interface {
	Serve(ctx context.Context) error
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
	httpController := http.NewShortenerController(
		serviceShortener,
		servicePinger,
		serviceAuth,
		serviceStats,
		conf)
	grpcController := grpcserver.NewGRPCController(
		conf,
		serviceShortener,
		servicePinger,
		serviceAuth,
		serviceStats,
	)

	startServer(ctx, httpController, grpcController)
}

func initRepo(ctx context.Context, conf config.Config) (shortenerRepoCloser, error) {
	if conf.DatabaseDsn != "" {
		logger.Log.Info("starting postgresql repo")

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
		logger.Log.Info("starting file repo")

		fileRepo, err := fileRepo.NewFileRepo(conf.FileStoragePath)
		if err != nil {
			return nil, err
		}

		return fileRepo, nil
	}

	logger.Log.Info("starting inmemory repo")
	return inmrRepo.NewInMemoryRepo(), nil
}

func startServer(ctx context.Context, httpController serverHandler, grpcController serverHandler) {
	var err error
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = httpController.Serve(ctx); err != nil {
			logger.Log.Error("http serve error: %s", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err = grpcController.Serve(ctx); err != nil {
			logger.Log.Error("grpc serve error: %s", zap.Error(err))
		}
	}()

	wg.Wait()
}
