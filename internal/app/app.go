package app

import (
	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	conf := config.New()
	serviceShortener := usecase.New()
	shortenerController := http.NewShortenerController(serviceShortener, conf)
	shortenerController.Serve()
}
