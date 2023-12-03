package app

import (
	"github.com/KartoonYoko/go-url-shortener/internal/controller/defaulthttp"
	"github.com/KartoonYoko/go-url-shortener/internal/usecase"
)

func Run() {
	var serviceShortener *usecase.Shortener = usecase.New()

	shortenerController := defaulthttp.NewShortenerController(serviceShortener)
	shortenerController.Serve()
}
