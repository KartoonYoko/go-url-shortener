package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	usecaseShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"
	"github.com/go-chi/chi/v5"
)

// Эндпоинт с методом POST и путём /.
// Сервер принимает в теле запроса строку URL как text/plain
// и возвращает ответ с кодом 201 и сокращённым URL как text/plain.
func (c *shortenerController) handlerRootPOST(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Empty body not allowed", http.StatusBadRequest)
		return
	}

	// - вернуть сокращенный url с помощью сервиса
	url, err := c.uc.SaveURL(ctx, string(body), userID)
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
func (c *shortenerController) handlerRootGET(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить id из строки запроса
	id := chi.URLParam(r, "id")
	// - получить из сервиса оригинальный url по id
	url, err := c.uc.GetURLByID(ctx, id)
	if err != nil {
		if errors.Is(err, usecaseShortener.ErrURLDeleted) {
			w.WriteHeader(http.StatusGone)
			return
		}
		http.Error(w, "Url not found", http.StatusBadRequest)
		return
	}

	// - в случае успеха вернуть 307 и url в заголовке "Location"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *shortenerController) handlerAPIShortenPOST(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request model.CreateShortenURLRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	if request.URL == "" {
		http.Error(w, "Empty body URL", http.StatusBadRequest)
		return
	}

	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Empty body not allowed", http.StatusBadRequest)
		return
	}

	url, err := c.uc.SaveURL(ctx, string(request.URL), userID)
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

func (c *shortenerController) handlerAPIShortenBatchPOST(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request []model.CreateShortenURLBatchItemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	response, err := c.uc.SaveURLsBatch(ctx, request, userID)
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

// Иметь хендлер GET /api/user/urls,
// который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
//
//			[
//	    		{
//	        		"short_url": "http://...",
//	        		"original_url": "http://..."
//	    		},
//	    		...
//			]
//
// При отсутствии сокращённых пользователем URL хендлер должен отдавать HTTP-статус 204 No Content.
func (c *shortenerController) handlerAPIUserURLsGET(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	response, err := c.uc.GetUserURLs(ctx, userID)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	if len(response) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Can not serialize response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseJSON))
}

func (c *shortenerController) handlerAPIUserURLsDELETE(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	var request []string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Can not parse body", http.StatusBadRequest)
		return
	}

	err = c.uc.DeleteURLs(ctx, userID, request)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *shortenerController) getUserIDFromContext(ctx context.Context) (string, error) {
	ctxUserID := ctx.Value(keyUserID)
	userID, ok := ctxUserID.(string)
	if !ok {
		msg := "can not get user ID from context"
		logger.Log.Debug(msg)
		return "", fmt.Errorf(msg)
	}

	return userID, nil
}
