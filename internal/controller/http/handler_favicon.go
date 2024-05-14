package http

import (
	"net/http"
)

// handlerFaviconGET нужен, чтобы не было ошибок в логах, когда открывается /debug/pprof:
// запрос иконки (/favicon.ico) совпадает с маршрутом получени сокращенного URL'a (/{id})
func (c *shortenerController) handlerFaviconGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Shortener"))
}
