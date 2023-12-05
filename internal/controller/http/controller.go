package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/config"
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
	r.Route("/", func(r chi.Router) {
		r.Get("/{id}", c.get)
		r.Post("/", c.post)
	})

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

func (c *shortenerController) Serve() {
	err := http.ListenAndServe(c.conf.BootstrapNetAddress, c.router)
	if err != nil {
		panic(err)
	}
}
