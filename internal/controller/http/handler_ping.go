package http

import (
	"context"
	"net/http"
)

func (c *ShortenerController) ping(w http.ResponseWriter, r *http.Request) {
	ctx := context.TODO()
	err := c.ucPing.Ping(ctx)
	if err != nil {
		http.Error(w, "Ping error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
