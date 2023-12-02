package app

import (
	"io"
	"net/http"
	"strings"

	"github.com/KartoonYoko/go-url-shortener/internal/service"
)

var serviceShortener *service.Shortener = service.New()

// Эндпоинт с методом POST и путём /.
// Сервер принимает в теле запроса строку URL как text/plain
// и возвращает ответ с кодом 201 и сокращённым URL как text/plain.
func post(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить из тела запроса строку
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Server error", http.StatusBadRequest)
		return
	}

	// - вернуть сокращенный url с помощью сервиса
	id := serviceShortener.SaveURL(string(body))

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(id))
}

// Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например, /EwHXdJfB).
// В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
func get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	// - получить id из строки запроса
	id := strings.TrimPrefix(r.URL.Path, "/")
	id, _ = strings.CutSuffix(id, "/")
	// - получить из сервиса оригинальный url по id
	url, err := serviceShortener.GetURLByID(id)
	if err != nil {
		http.Error(w, "Not found key", http.StatusBadRequest)
		return
	}

	// - в случае успеха вернуть 307 и url в заголовке "Location"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			post(w, r)
		} else if r.Method == http.MethodGet {
			get(w, r)
		} else {
			http.Error(w, "Method is not allowed", http.StatusBadRequest)
			return
		}

		if next == nil {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Serve() {
	mux := http.NewServeMux()

	mux.Handle("/", middleware(nil))
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
