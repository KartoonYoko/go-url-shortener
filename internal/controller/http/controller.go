package http

import (
	"context"
	"log"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/go-chi/chi/v5"
)

type useCaseShortener interface {
	GetURLByID(id string) (string, error)
	SaveURL(url string) (string, error)
}

type useCasePinger interface {
	Ping(ctx context.Context) error
}

type shortenerController struct {
	uc     useCaseShortener
	ucPing useCasePinger
	router *chi.Mux
	conf   *config.Config
}

func NewShortenerController(uc useCaseShortener, ucPing useCasePinger,
	conf *config.Config) *shortenerController {
	c := &shortenerController{
		uc:     uc,
		ucPing: ucPing,
		conf:   conf,
	}
	r := chi.NewRouter()

	// middlewares
	r.Use(logRequestTimeMiddleware)
	r.Use(decompressRequestGZIPMiddleware)
	r.Use(compressResponseGZIPMiddleware)
	r.Use(logResponseInfoMiddleware)

	// routes
	// root
	{
		rootRouter := chi.NewRouter()
		rootRouter.Route("/", func(r chi.Router) {
			r.Get("/{id}", c.get)
			r.Post("/", c.post)
		})

		r.Mount("/", rootRouter)
	}

	// api
	{
		apiRouter := chi.NewRouter()
		apiRouter.Route("/", func(r chi.Router) {
			r.Post("/shorten", c.postCreateShorten)
		})

		r.Mount("/api", apiRouter)
	}

	// ping
	{
		pingRouter := chi.NewRouter()
		pingRouter.Route("/", func(r chi.Router) {
			r.Get("/", c.ping)
		})

		r.Mount("/ping", pingRouter)
	}

	c.router = r
	return c
}

func (c *shortenerController) Serve() {
	err := http.ListenAndServe(c.conf.BootstrapNetAddress, c.router)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
