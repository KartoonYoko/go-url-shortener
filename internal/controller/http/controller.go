package http

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

type useCase interface {
	GetURLByID(id string) (string, error)
	SaveURL(url string) string
}

type shortenerController struct {
	uc     useCase
	router *chi.Mux
	conf   *config.Config
}

func NewShortenerController(uc useCase, conf *config.Config) *shortenerController {
	c := &shortenerController{
		uc:   uc,
		conf: conf,
	}

	r := chi.NewRouter()

	r.Use(logRequestTimeMiddleware)
	r.Use(logResponseInfoMiddleware)

	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", c.get)
		r.Post("/", c.post)
	})

	apiRouter := chi.NewRouter()
	apiRouter.Route("/", func(r chi.Router) {
		r.Post("/shorten", c.postCreateShorten)
	})

	r.Mount("/api", apiRouter)
	c.router = r

	return c
}

// Эндпоинт с методом POST и путём /.
// Сервер принимает в теле запроса строку URL как text/plain
// и возвращает ответ с кодом 201 и сокращённым URL как text/plain.
func (c *shortenerController) post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить из тела запроса строку
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Server error", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty body not allowed", http.StatusBadRequest)
		return
	}
	// - вернуть сокращенный url с помощью сервиса
	id := c.uc.SaveURL(string(body))

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	res := fmt.Sprintf("%s/%s", c.conf.BaseURLAddress, id)
	w.Write([]byte(res))
}

// Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например, /EwHXdJfB).
// В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (c *shortenerController) get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить id из строки запроса
	id := chi.URLParam(r, "id")
	// - получить из сервиса оригинальный url по id
	url, err := c.uc.GetURLByID(id)
	if err != nil {
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}

	// - в случае успеха вернуть 307 и url в заголовке "Location"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *shortenerController) postCreateShorten(w http.ResponseWriter, r *http.Request) {
	var request model.CreateShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	if request.Url == "" {
		http.Error(w, "Empty body URL", http.StatusBadRequest)
		return
	}

	id := c.uc.SaveURL(string(request.Url))

	response := model.CreateShortenURLResponse{
		Result: fmt.Sprintf("%s/%s", c.conf.BaseURLAddress, id),
	}
	res, err := json.Marshal(&response)
	if err != nil {
		http.Error(w, "Can not serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res))
}

func (c *shortenerController) Serve() {
	err := http.ListenAndServe(c.conf.BootstrapNetAddress, c.router)
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
