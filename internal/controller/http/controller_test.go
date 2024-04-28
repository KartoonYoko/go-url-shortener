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
	"github.com/KartoonYoko/go-url-shortener/internal/repository"
	inmr "github.com/KartoonYoko/go-url-shortener/internal/repository/inmemoryrepo"
	ucShortener "github.com/KartoonYoko/go-url-shortener/internal/usecase/shortener"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	controller *shortenerController
	srv        *httptest.Server
	ucMock     *useCaseMock
)

func TestMain(m *testing.M) {
	controller = createTestMock()
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv = httptest.NewServer(controller.router)
	// останавливаем сервер после завершения теста
	defer srv.Close()
	controller.conf.BaseURLAddress = srv.URL

	m.Run()
}

func TearDownTest(t *testing.T) {
	require.NoError(t, ucMock.Clear(context.Background()))
}

// createTestMock собирает контроллер
// Пока непонятно как правильно инициализировать данные, поэтому пока так.
func createTestMock() *shortenerController {
	ucMock = &useCaseMock{
		repo:           *inmr.NewInMemoryRepo(),
		baseAddressURL: "http://127.0.0.1:8080", // задаём любой URL, который попадёт под регулярку в тестах
	}
	c := NewShortenerController(ucMock, ucMock, ucMock, &config.Config{})
	return c
}

type useCaseMock struct {
	repo           inmr.InMemoryRepo
	baseAddressURL string
}

func (s *useCaseMock) SaveURL(ctx context.Context, url string, userID string) (string, error) {
	id, err := s.repo.SaveURL(ctx, url, userID)
	if err != nil {
		var repoErrURLAlreadyExists *repository.URLAlreadyExistsError
		if errors.As(err, &repoErrURLAlreadyExists) {
			return "", ucShortener.NewURLAlreadyExistsError(repoErrURLAlreadyExists.ID, repoErrURLAlreadyExists.URL, err)
		}

		return "", err
	}
	return id, nil
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

func (s *useCaseMock) DeleteURLs(ctx context.Context, userID string, urlsIDs []string) error {
	return nil
}

func (s *useCaseMock) Clear(ctx context.Context) error {
	return s.repo.Clear()
}

func (s *useCaseMock) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func TestPost(t *testing.T) {
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
		{
			name: "Already exists",
			url:  "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
			want: want{
				code:          http.StatusConflict,
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

	TearDownTest(t)
}

func TestPostAPIShorten(t *testing.T) {
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
			name: "Already exists",
			request: model.CreateShortenURLRequest{
				URL: "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
			},
			want: want{
				code:          http.StatusConflict,
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

	TearDownTest(t)
}

func TestPostAPIShortenBatch(t *testing.T) {
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

	TearDownTest(t)
}

func TestGet(t *testing.T) {
	ctx := context.TODO()

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

	TearDownTest(t)
}

func TestHandlerAPIUserURLsGET(t *testing.T) {
	apiRoute := "/api/user/urls"

	// какой результат хотим получить
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		prepare func (t *testing.T, httpClient *resty.Client)
		want want
	}{
		{
			name: "No content",
			want: want{
				code:        http.StatusNoContent,
				contentType: "",
			},
		},
		{
			name: "Get content",
			prepare: func (t *testing.T, httpClient *resty.Client) {
				createURL(t, "https://pkg.go.dev/regexp#example-Match", httpClient)
			},
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

	// авторизируемся
	auth(t, jar)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.prepare != nil {
				test.prepare(t, httpClient)
			}
			res, err := httpClient.R().Get(srv.URL + apiRoute)
			require.NoError(t, err)
			assert.Equal(t, test.want.code, res.StatusCode())
			if test.want.contentType != "" {
				assert.Contains(t, test.want.contentType, res.Header().Get("Content-Type"))
			}
		})
	}

	TearDownTest(t)
}

func TestHandlerAPIUserURLsDELETE(t *testing.T) {
	apiRoute := "/api/user/urls"

	httpClient := resty.
		New().
		SetBaseURL(srv.URL)
	request := []string{
		"https://angular.io",
		"https://www.postgresqltutorial.com",
	}
	res, err := httpClient.
		R().
		SetBody(request).
		Delete(srv.URL + apiRoute)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode())

	TearDownTest(t)
}

func createURL(t *testing.T, url string, httpClient *resty.Client) {
	req := httpClient.
		R().
		SetBody(url)
	resp, err := req.Post("/")
	assert.NoError(t, err, "Ошибка при попытке сделать запрос для сокращения URL")

	assert.Equalf(t, http.StatusCreated, resp.StatusCode(),
		"Несоответствие статус кода ответа ожидаемому в хендлере '%s %s'", req.Method, req.URL)

}

func auth(t *testing.T, jar *cookiejar.Jar) {
	userID, err := ucMock.GetNewUserID(context.Background())
	require.NoError(t, err)

	pURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	jwt, err := buildJWTString(userID)
	require.NoError(t, err)
	bearerStr := fmt.Sprintf("Bearer %s", jwt)
	cookie := createAuthCookie(bearerStr)
	jar.SetCookies(pURL, []*http.Cookie{&cookie})
}
