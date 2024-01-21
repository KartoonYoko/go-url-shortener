package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/internal/model"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"
	"github.com/go-chi/chi/v5"
)

// Эндпоинт с методом POST и путём /.
// Сервер принимает в теле запроса строку URL как text/plain
// и возвращает ответ с кодом 201 и сокращённым URL как text/plain.
func (c *shortenerController) post(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
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
	url, err := c.uc.SaveURL(ctx, string(body))
	if err != nil {
		var alreadyExistsErr *usecaseShortener.URLAlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			w.Header().Set("content-type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(alreadyExistsErr.ShortURL))
			return
		}
		http.Error(w, "Server error", http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(url))
}

// Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например, /EwHXdJfB).
// В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func (c *shortenerController) get(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить id из строки запроса
	id := chi.URLParam(r, "id")
	// - получить из сервиса оригинальный url по id
	url, err := c.uc.GetURLByID(ctx, id)
	if err != nil {
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}

	// - в случае успеха вернуть 307 и url в заголовке "Location"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *shortenerController) postCreateShorten(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()

	var request model.CreateShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "Empty body URL", http.StatusBadRequest)
		return
	}

	url, err := c.uc.SaveURL(ctx, string(request.URL))
	if err != nil {
		var alreadyExistsErr *usecaseShortener.URLAlreadyExistsError
		if errors.As(err, &alreadyExistsErr) {
			res, err := json.Marshal(model.CreateShortenURLResponse{
				Result: alreadyExistsErr.ShortURL,
			})

			if err != nil {
				http.Error(w, "Can not serialize response", http.StatusInternalServerError)
				return
			}

			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(res))
			return
		}
		http.Error(w, "Server error", http.StatusBadRequest)
		return
	}

	response := model.CreateShortenURLResponse{
		Result: url,
	}
	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Can not serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(res))
}

func (c *shortenerController) postCreateShortenBatch(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	var request []model.CreateShortenURLBatchItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	response, err := c.uc.SaveURLsBatch(ctx, request)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Can not serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(responseJSON))
}
