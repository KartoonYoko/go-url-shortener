package http

import (
	"encoding/json"
	"net/http"
)

func (c *shortenerController) handlerStatsGET(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	st, err := c.ucStats.GetStats(ctx)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(st)

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(res))
}
