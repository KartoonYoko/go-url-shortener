package app

import (
	"database/sql"
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	repository "github.com/KartoonYoko/go-url-shortener/internal/repository/shortener"
	usecasePinger "github.com/KartoonYoko/go-url-shortener/internal/usecase/ping"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Run() {
	// logger
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()
	conf := config.New()

	// репозитории
	db, err := sql.Open("pgx", conf.DatabaseDsn)
	if err != nil {
		log.Fatal(err)
	}
	repo, err := repository.NewPsgsqlRepo(db)
	if err != nil {
		log.Fatal(err)
	}
	defer repo.Close()
	fileRepo, err := repository.NewFileRepo(conf.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileRepo.Close()

	// usecase'ы
	serviceShortener := usecaseShortener.New(fileRepo)
	servicePinger := usecasePinger.NewPingUseCase(repo)

	// контроллеры
	shortenerController := http.NewShortenerController(serviceShortener, servicePinger, conf)

	shortenerController.Serve()
}
