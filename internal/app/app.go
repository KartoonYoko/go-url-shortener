package app

import (
	"log"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	if err := logger.Initialize("Info"); err != nil {
		log.Fatal(err)
	}
	defer logger.Log.Sync()

	conf := config.New()
	serviceShortener := usecase.New()
	shortenerController := http.NewShortenerController(serviceShortener, conf)
	shortenerController.Serve()
}
