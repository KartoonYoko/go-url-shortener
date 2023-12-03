package defaulthttp

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type useCase interface {
	GetURLByID(id string) (string, error)
	SaveURL(url string) string
}

type shortenerController struct {
	uc useCase
}

func NewShortenerController(uc useCase) *shortenerController {
	return &shortenerController{
		uc: uc,
	}
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
	res := fmt.Sprintf("http://%s/%s", r.Host, id)
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
	id := strings.TrimPrefix(r.URL.Path, "/")
	id, _ = strings.CutSuffix(id, "/")
	// - получить из сервиса оригинальный url по id
	url, err := c.uc.GetURLByID(id)
	if err != nil {
		http.Error(w, "Not found key", http.StatusBadRequest)
		return
	}

	// - в случае успеха вернуть 307 и url в заголовке "Location"
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c *shortenerController) router(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		c.post(w, r)
	} else if r.Method == http.MethodGet {
		c.get(w, r)
	} else {
		http.Error(w, "Method is not allowed", http.StatusBadRequest)
		return
	}
}

func (c *shortenerController) Serve() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", c.router)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
