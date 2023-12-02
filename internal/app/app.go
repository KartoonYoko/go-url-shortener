package app

import (
	"fmt"
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
	res := fmt.Sprintf("http://%s/%s", r.Host, id)
	w.Write([]byte(res))
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

func router(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		post(w, r)
	} else if r.Method == http.MethodGet {
		get(w, r)
	} else {
		http.Error(w, "Method is not allowed", http.StatusBadRequest)
		return
	}
}

func Serve() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", router)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
