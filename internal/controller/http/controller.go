package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type useCaseShortener interface {
	GetURLByID(ctx context.Context, urlID string) (string, error)
	SaveURL(ctx context.Context, url string, userID string) (string, error)
	SaveURLsBatch(ctx context.Context,
		request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error)
	GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error)
	DeleteURLs(ctx context.Context, userID string, urlsIDs []string) error
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

// NewShortenerController собирает http контроллер, определяя endpoint'ы, middleware'ы
func NewShortenerController(
	uc useCaseShortener,
	ucPing useCasePinger,
	ucAuth useCaseAuther,
	conf *config.Config) *shortenerController {
	c := &shortenerController{
		uc:     uc,
		ucAuth: ucAuth,
		ucPing: ucPing,
		conf:   conf,
	}
	r := chi.NewRouter()

	// middlewares
	r.Use(middleware.Logger)
	r.Use(logRequestTimeMiddleware)
	r.Use(decompressRequestGZIPMiddleware)
	r.Use(c.authJWTCookieMiddleware)
	r.Use(compressResponseGZIPMiddleware)
	r.Use(logResponseInfoMiddleware)

	// routes
	r.Mount("/debug", middleware.Profiler())
	routeRoot(r, c)
	routeAPI(r, c)
	routePing(r, c)

	c.router = r
	return c
}

func routeRoot(r *chi.Mux, c *shortenerController) {
	r.Get("/favicon.ico", c.handlerFaviconGET)
	r.Get("/{id}", c.handlerRootGET)
	r.Post("/", c.handlerRootPOST)
}

func routeAPI(r *chi.Mux, c *shortenerController) {
	apiRouter := chi.NewRouter()
	apiRouter.Post("/shorten", c.handlerAPIShortenPOST)
	apiRouter.Post("/shorten/batch", c.handlerAPIShortenBatchPOST)
	apiRouter.Get("/user/urls", c.handlerAPIUserURLsGET)
	apiRouter.Delete("/user/urls", c.handlerAPIUserURLsDELETE)

	r.Mount("/api", apiRouter)
}

func routePing(r *chi.Mux, c *shortenerController) {
	pingRouter := chi.NewRouter()
	pingRouter.Get("/", c.ping)

	r.Mount("/ping", pingRouter)
}

// Serve запускает http сервер
func (c *shortenerController) Serve(ctx context.Context) {
	server := &http.Server{Addr: c.conf.BootstrapNetAddress, Handler: c.router}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(ctx)

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	// run server
	logger.Log.Info(fmt.Sprintf("server serve on %s", c.conf.BootstrapNetAddress))
	err := http.ListenAndServe(c.conf.BootstrapNetAddress, c.router)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
