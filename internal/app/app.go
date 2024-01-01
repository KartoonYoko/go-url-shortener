package app

import (
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"github.com/KartoonYoko/go-url-shortener/internal/repository"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()

	conf := config.New()

	repo, err := repository.NewFileRepo(conf.FileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	serviceShortener := usecase.New(repo)
	shortenerController := http.NewShortenerController(serviceShortener, conf)
	shortenerController.Serve()
}
