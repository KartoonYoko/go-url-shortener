package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/KartoonYoko/go-url-shortener/config"
	model "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	repository "github.com/KartoonYoko/go-url-shortener/internal/repository/shortener"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type useCaseMock struct {
	repo repository.InMemoryRepo
	// storage        map[string]string
	// r              *rand.Rand
	// letterRunes    []rune
	baseAddressURL string
}

func (s *useCaseMock) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	return s.repo.SaveURL(ctx, url, userID)
}

func (s *useCaseMock) GetURLByID(ctx context.Context, id string) (string, error) {
	return s.repo.GetURLByID(ctx, id)
}

func (s *useCaseMock) GetUserURLs(ctx context.Context, userID string) ([]model.GetUserURLsItemResponse, error) {
	return s.repo.GetUserURLs(ctx, userID)
}

func (s *useCaseMock) SaveURLsBatch(ctx context.Context,
	request []model.CreateShortenURLBatchItemRequest, userID string) ([]model.CreateShortenURLBatchItemResponse, error) {
	return s.repo.SaveURLsBatch(ctx, request, userID)
}

func (s *useCaseMock) GetNewUserID(ctx context.Context) (string, error) {
	return s.repo.GetNewUserID(ctx)
}

// Метод собирает нужный контроллер, нужно вызывать в каждой функции.
// Пока непонятно как правильно инициализировать данные, поэтому пока так.
func createTestMock() *shortenerController {
	uc := &useCaseMock{
		repo: *repository.NewInMemoryRepo(),
		// r:              rand.New(rand.NewSource(time.Now().UnixMilli())),
		// storage:        make(map[string]string),
		// letterRunes:    []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		baseAddressURL: "http://127.0.0.1:8080", // задаём любой URL, который попадёт под регулярку в тестах
	}
	c := NewShortenerController(uc, nil, uc, &config.Config{})
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
	urlRegex := `[A-z0-9]+`
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
	urlRegex := `[A-z0-9]+`
	tests := []struct {
		name    string
		request model.CreateShortenURLRequest
		want    want
	}{
		{
			name: "Simple positive request",
			request: model.CreateShortenURLRequest{
				URL: "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
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
				URL: "",
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

func TestPostAPIShortenBatch(t *testing.T) {
	controller := createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL
	apiRoute := "/api/shorten/batch"

	// какой результат хотим получить
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name    string
		request []model.CreateShortenURLBatchItemRequest
		want    want
	}{
		{
			name: "Different URLs",
			request: []model.CreateShortenURLBatchItemRequest{
				{
					OriginalURL:   "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
					CorrelationID: "0",
				},
				{
					OriginalURL:   "https://github.com/docker/compose",
					CorrelationID: "1",
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name: "Existing URLs",
			request: []model.CreateShortenURLBatchItemRequest{
				{
					OriginalURL:   "https://github.com/docker/compose",
					CorrelationID: "0",
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name: "Repetitive and existing URLs",
			request: []model.CreateShortenURLBatchItemRequest{
				{
					OriginalURL:   "https://github.com/docker/compose",
					CorrelationID: "0",
				},
				{
					OriginalURL:   "https://github.com/nasa/astrobee",
					CorrelationID: "1",
				},
				{
					OriginalURL:   "https://github.com/docker/compose",
					CorrelationID: "2",
				},
			},
			want: want{
				code:        http.StatusCreated,
				contentType: "application/json",
			},
		},
	}

	errRedirectBlocked := errors.New("HTTP redirect blocked")
	redirPolicy := resty.RedirectPolicyFunc(func(_ *http.Request, _ []*http.Request) error {
		return errRedirectBlocked
	})
	httpClient := resty.New().SetBaseURL(srv.URL).SetRedirectPolicy(redirPolicy)
	urlRegex := `[A-z0-9]+`
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpClient.R().SetBody(test.request).Post(srv.URL + apiRoute)
			require.NoError(t, err)
			assert.Equal(t, test.want.code, res.StatusCode())
			assert.Contains(t, res.Header().Get("Content-Type"), test.want.contentType)

			var response []model.CreateShortenURLBatchItemResponse
			err = json.Unmarshal(res.Body(), &response)
			require.NoError(t, err)
			assert.Equal(t, len(response), len(test.request), "not equals count of response and request")

			for _, res := range response {
				assert.Regexpf(t, urlRegex, res.ShortURL, "Result short URL (%s) is not matched regex (%s)", res.ShortURL, urlRegex)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.TODO()

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
		urc.urlID, _ = controller.uc.SaveURL(ctx, urc.url, "some user id")
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

func TestHandlerAPIUserURLsGET(t *testing.T) {
	controller := createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL
	apiRoute := "/api/user/urls"
	unauthorizedTestName := "Unauthorized"

	// какой результат хотим получить
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: unauthorizedTestName,
			want: want{
				code:        http.StatusUnauthorized,
				contentType: "",
			},
		},
		{
			name: "Get content",
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
	}

	// создаем cookie jar для сохранения cookies между запросами
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	httpClient := resty.
		New().
		SetBaseURL(srv.URL).
		SetCookieJar(jar)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := httpClient.R().Get(srv.URL + apiRoute)
			require.NoError(t, err)
			assert.Equal(t, test.want.code, res.StatusCode())
			if test.want.contentType != "" {
				assert.Contains(t, res.Header().Get("Content-Type"), test.want.contentType)
			}

			// если неавторизованный тест, то пробуем добавить URL для пользователя
			if test.name == unauthorizedTestName {
				req := httpClient.
					R().
					SetBody("https://music.yandex.ru/home")
				resp, err := req.Post("/")
				assert.NoError(t, err, "Ошибка при попытке сделать запрос для сокращения URL")
				shortenURL := string(resp.Body())
				assert.Equalf(t, http.StatusCreated, resp.StatusCode(),
					"Несоответствие статус кода ответа ожидаемому в хендлере '%s %s'", req.Method, req.URL)

				_, urlParseErr := url.Parse(shortenURL)
				assert.NoErrorf(t, urlParseErr,
					"Невозможно распарсить полученный сокращенный URL - %s : %s", shortenURL, err,
				)
			}
		})
	}
}
