package http

import (
	"context"
	"log"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/go-chi/chi/v5"
)

type useCaseShortener interface {
	GetURLByID(ctx context.Context, id string) (string, error)
	SaveURL(ctx context.Context, url string) (string, error)
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
	routeRoot(r, c)
	routeAPI(r, c)
	routePing(r, c)

	c.router = r
	return c
}

func routeRoot(r *chi.Mux, c *shortenerController) {
	r.Get("/{id}", c.get)
	r.Post("/", c.post)
}

func routeAPI(r *chi.Mux, c *shortenerController) {
	apiRouter := chi.NewRouter()
	apiRouter.Post("/shorten", c.postCreateShorten)

	r.Mount("/api", apiRouter)
}

func routePing(r *chi.Mux, c *shortenerController) {
	pingRouter := chi.NewRouter()
	pingRouter.Get("/", c.ping)

	r.Mount("/ping", pingRouter)
}

func (c *shortenerController) Serve() {
	err := http.ListenAndServe(c.conf.BootstrapNetAddress, c.router)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
