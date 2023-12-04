package app

import (
	"github.com/KartoonYoko/go-url-shortener/internal/controller/http"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	serviceShortener := usecase.New()

	shortenerController := http.NewShortenerController(serviceShortener)
	shortenerController.Serve()
}
