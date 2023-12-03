package app

import (
	"github.com/KartoonYoko/go-url-shortener/internal/controller/defaulthttp"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	serviceShortener := usecase.New()

	shortenerController := defaulthttp.NewShortenerController(serviceShortener)
	shortenerController.Serve()
}
