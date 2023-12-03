package defaulthttp

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	return NewShortenerController(uc)
}

func TestPost(t *testing.T) {
	controller := createTestMock()
	// какой результат хотим получить
	type want struct {
		code          int
		responseRegex string
		contentType   string
	}

	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Post test #1",
			url:  "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96",
			want: want{
				code:          http.StatusCreated,
				responseRegex: `http://localhost:8080/[A-z0-9]*`,
				contentType:   "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.url))
			w := httptest.NewRecorder()
			controller.post(w, request)

			res := w.Result()
			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// проверяем тип контента
			assert.Contains(t, res.Header.Get("Content-Type"), test.want.contentType)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.NotRegexpf(t, test.url, resBody, "Body result (%s) is not matched regex (%s)", resBody, test.url)
		})
	}
}

func TestGet(t *testing.T) {
	controller := createTestMock()
	type want struct {
		code int
	}
	// короткие ссылки полученные от сервиса
	type useCaseURLCheck struct {
		urlId string
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
				urlId: "urlIdThatNotExists",
				url:   "",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	urlsToCheck := []useCaseURLCheck{
		{urlId: "", url: "https://pkg.go.dev/regexp#example-Match"},
		{urlId: "", url: "https://gist.github.com/brydavis/0c7da92bd508195744708eeb2b54ac96"},
	}
	for i, urc := range urlsToCheck {
		urc.urlId = controller.uc.SaveURL(urc.url)
		tests = append(tests, testData{
			name:    fmt.Sprintf("Positive request #%d", i),
			urlData: urc,
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/"+test.urlData.urlId, nil)
			w := httptest.NewRecorder()
			controller.get(w, request)
			res := w.Result()

			// проверяем код ответа
			assert.Equal(t, test.want.code, res.StatusCode)
			// проверяем url
			assert.Equal(t, test.urlData.url, res.Header.Get("Location"))
		})
	}
}
