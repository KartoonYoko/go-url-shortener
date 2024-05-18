package http

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_shortenerController_handlerStatsGET(t *testing.T) {
	httpClient := resty.New().
		SetBaseURL(srv.URL)

	res, err := httpClient.R().Get("/api/internal/stats")

	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, res.StatusCode())
}
