package http

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	modelStats "github.com/KartoonYoko/go-url-shortener/internal/model/stats"
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

type useCaseStats interface {
	GetStats(ctx context.Context) (*modelStats.StatsResponse, error)
}

type ShortenerController struct {
	uc      useCaseShortener
	ucPing  useCasePinger
	ucAuth  useCaseAuther
	ucStats useCaseStats
	router  *chi.Mux
	conf    *config.Config
}

// NewShortenerController собирает http контроллер, определяя endpoint'ы, middleware'ы
func NewShortenerController(
	uc useCaseShortener,
	ucPing useCasePinger,
	ucAuth useCaseAuther,
	ucStats useCaseStats,
	conf *config.Config) *ShortenerController {
	c := &ShortenerController{
		uc:      uc,
		ucAuth:  ucAuth,
		ucPing:  ucPing,
		ucStats: ucStats,
		conf:    conf,
	}
	r := chi.NewRouter()

	// middlewares
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

func routeRoot(r *chi.Mux, c *ShortenerController) {
	r.Get("/favicon.ico", c.handlerFaviconGET)
	r.Get("/{id}", c.handlerRootGET)
	r.Post("/", c.handlerRootPOST)
}

func routeAPI(r *chi.Mux, c *ShortenerController) {
	apiRouter := chi.NewRouter()

	apiRouter.Group(func(r chi.Router) {
		r.Post("/shorten", c.handlerAPIShortenPOST)
		r.Post("/shorten/batch", c.handlerAPIShortenBatchPOST)
	})

	apiRouter.Group(func(r chi.Router) {
		r.Get("/user/urls", c.handlerAPIUserURLsGET)
		r.Delete("/user/urls", c.handlerAPIUserURLsDELETE)
	})

	apiRouter.Group(func(r chi.Router) {
		r.Use(c.guardIPMiddleware)

		r.Get("/internal/stats", c.handlerStatsGET)
	})

	r.Mount("/api", apiRouter)
}

func routePing(r *chi.Mux, c *ShortenerController) {
	pingRouter := chi.NewRouter()
	pingRouter.Get("/", c.ping)

	r.Mount("/ping", pingRouter)
}

// Serve запускает http сервер
func (c *ShortenerController) Serve(ctx context.Context) error {
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
	logger.Log.Info(fmt.Sprintf("server serve on %s", server.Addr))
	if c.conf.EnableHTTPS {
		cert, key, err := createCert()
		if err != nil {
			return fmt.Errorf("create cert error: %w", err)
		}
		err = server.ListenAndServeTLS(cert, key)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server serve error: %w", err)
		}
	} else {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server serve error: %w", err)
		}
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()

	return nil
}

func createCert() (certPath string, keyPath string, err error) {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"localhost"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	var privateKeyPEM bytes.Buffer
	pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile("cert.pem", certPEM.Bytes(), 0644)
	if err != nil {
		return
	}
	err = os.WriteFile("key.pem", privateKeyPEM.Bytes(), 0600)
	if err != nil {
		return
	}

	certPath = "cert.pem"
	keyPath = "key.pem"
	return
}
