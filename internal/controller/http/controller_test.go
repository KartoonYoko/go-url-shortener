package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/KartoonYoko/go-url-shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type useCaseMock struct {
	storage     map[string]string
	r           *rand.Rand
	letterRunes []rune
}

func (s *useCaseMock) SaveURL(url string) string {
	hash := s.randStringRunes(5)
	s.storage[hash] = url
	return hash
}

func (s *useCaseMock) GetURLByID(id string) (string, error) {
	res := s.storage[id]

	if res == "" {
		return res, fmt.Errorf("Not found url by id %s", id)
	}

	return res, nil
}

func (s *useCaseMock) randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = s.letterRunes[rand.Intn(len(s.letterRunes))]
	}
	return string(b)
}

// Метод собирает нужный контроллер, нужно вызывать в каждой функции.
// Пока непонятно как правильно инициализировать данные, поэтому пока так.
func createTestMock() *shortenerController {
	uc := &useCaseMock{
		r:           rand.New(rand.NewSource(time.Now().UnixMilli())),
		storage:     make(map[string]string),
		letterRunes: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
	}
	c := NewShortenerController(uc, &config.Config{})
	return c
}

func TestPost(t *testing.T) {
	controller := createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL

	// какой результат хотим получить
	type want struct {
		code          int
		responseRegex string
		contentType   string
	}

	// под данную регулярку должны попадать ответы сервера
	urlRegex := `http:\/\/[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{2,5}\/[A-z0-9]+`
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Simple positive request",
			url:  "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
			want: want{
				code:          http.StatusCreated,
				responseRegex: urlRegex,
				contentType:   "text/plain",
			},
		},
		{
			name: "Empty body request",
			url:  "",
			want: want{
				code:          http.StatusBadRequest,
				responseRegex: urlRegex,
				contentType:   "text/plain",
			},
		},
	}

	httpClient := resty.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpClient.R().SetBody(test.url).Post(srv.URL)
			require.NoError(t, err)
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode())
			// проверяем тип контента
			assert.Contains(t, res.Header().Get("Content-Type"), test.want.contentType)

			if test.want.code == http.StatusCreated {
				// получаем и проверяем тело запроса
				resBody := res.Body()
				body := string(resBody)
				assert.Regexpf(t, test.want.responseRegex, body, "Body result (%s) is not matched regex (%s)", body, test.want.responseRegex)
			}
		})
	}
}

func TestPostAPIShorten(t *testing.T) {
	controller := createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL
	apiRoute := "/api/shorten"

	// какой результат хотим получить
	type want struct {
		code          int
		responseRegex string
		contentType   string
	}

	// под данную регулярку должны опападать ответы сервера
	urlRegex := `http:\/\/[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}:[0-9]{2,5}\/[A-z0-9]+`
	tests := []struct {
		name    string
		request model.CreateShortenURLRequest
		want    want
	}{
		{
			name: "Simple positive request",
			request: model.CreateShortenURLRequest{
				Url: "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
			},
			want: want{
				code:          http.StatusCreated,
				responseRegex: urlRegex,
				contentType:   "application/json",
			},
		},
		{
			name: "Empty body request",
			request: model.CreateShortenURLRequest{
				Url: "",
			},
			want: want{
				code:          http.StatusBadRequest,
				responseRegex: urlRegex,
				contentType:   "text/plain",
			},
		},
	}

	httpClient := resty.New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpClient.R().SetBody(test.request).Post(srv.URL + apiRoute)
			require.NoError(t, err)
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode())
			// проверяем тип контента
			assert.Contains(t, res.Header().Get("Content-Type"), test.want.contentType)

			if test.want.code == http.StatusCreated {
				// получаем и проверяем тело запроса
				var response model.CreateShortenURLResponse
				json.Unmarshal(res.Body(), &response)
				assert.Regexpf(t, test.want.responseRegex, response.Result, "Body result (%s) is not matched regex (%s)", response.Result, test.want.responseRegex)
			}
		})
	}
}

func TestGet(t *testing.T) {
	controller := createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL

	type want struct {
		code int
	}
	// короткие ссылки полученные от сервиса
	type useCaseURLCheck struct {
		urlID string
		url   string
	}
	type testData struct {
		name    string
		urlData useCaseURLCheck
		want    want
	}

	tests := []testData{
		{
			name: "Bad request #1",
			urlData: useCaseURLCheck{
				urlID: "urlIdThatNotExists",
				url:   "",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	urlsToCheck := []useCaseURLCheck{
		{urlID: "", url: "https://pkg.go.dev/regexp#example-Match"},
		{urlID: "", url: "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96"},
	}
	for i, urc := range urlsToCheck {
		urc.urlID = controller.uc.SaveURL(urc.url)
		tests = append(tests, testData{
			name:    fmt.Sprintf("Positive request #%d", i+1),
			urlData: urc,
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		})
	}

	// создаем HTTP клиент без поддержки редиректов
	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})
	httpClient := resty.New().
		SetBaseURL(srv.URL).
		SetRedirectPolicy(redirPolicy)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpClient.R().Get(test.urlData.urlID)
			if !errors.Is(err, errRedirectBlocked) {
				require.NoError(t, err)
			}

			// проверяем код ответа
			statusErr := assert.Equal(t, test.want.code, res.StatusCode())
			// проверяем url
			urlErr := assert.Equal(t, test.urlData.url, res.Header().Get("Location"))

			if !statusErr || !urlErr {
				t.Logf("Requested url: %s", res.Request.URL)
				t.Logf("Requesst method: %s", res.Request.Method)
				t.Logf("Body: %s", res.Body())
			}
		})
	}
}
