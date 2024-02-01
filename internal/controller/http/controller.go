package http

import (
	"context"
	"log"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/config"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/go-chi/chi/v5"
)

type useCaseShortener interface {
	GetURLByID(ctx context.Context, urlID string) (string, error)
	SaveURL(ctx context.Context, url string, userID string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error)
}

type useCasePinger interface {
	Ping(ctx context.Context) error
}

type useCaseAuther interface {
	GetNewUserID(ctx context.Context) (string, error)
}

type shortenerController struct {
	uc     useCaseShortener
	ucPing useCasePinger
	ucAuth useCaseAuther
	router *chi.Mux
	conf   *config.Config
}

func NewShortenerController(uc useCaseShortener, ucPing useCasePinger, ucAuth useCaseAuther,
	conf *config.Config) *shortenerController {
	c := &shortenerController{
		uc:     uc,
		ucAuth: ucAuth,
		ucPing: ucPing,
		conf:   conf,
	}
	r := chi.NewRouter()

	// middlewares
	r.Use(logRequestTimeMiddleware)
	r.Use(decompressRequestGZIPMiddleware)
	r.Use(c.authJWTCookieMiddleware)
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
	r.Get("/{id}", c.handlerRootGET)
	r.Post("/", c.handlerRootPOST)
}

func routeAPI(r *chi.Mux, c *shortenerController) {
	apiRouter := chi.NewRouter()
	apiRouter.Post("/shorten", c.handlerAPIShortenPOST)
	apiRouter.Post("/shorten/batch", c.handlerAPIShortenBatchPOST)
	apiRouter.Get("/user/urls", c.handlerAPIUserURLsGET)

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
