package http

import (
	"context"
	"time"

	"github.com/KartoonYoko/go-url-shortener/config"
	inmr "github.com/KartoonYoko/go-url-shortener/internal/repository/inmemoryrepo"
)

func Example() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	uc := &useCaseMock{
		repo: *inmr.NewInMemoryRepo(),
		baseAddressURL: "http://127.0.0.1:8080", // задаём любой URL, который попадёт под регулярку в тестах
	}
	c := NewShortenerController(uc, nil, uc, nil, &config.Config{})

	c.Serve(ctx)
}
