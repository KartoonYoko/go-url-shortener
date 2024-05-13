package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/KartoonYoko/go-url-shortener/internal/controller/common"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"go.uber.org/zap"
)

// MiddlewareAuthKey тип ключа контекста для middleware аутентификации
type MiddlewareAuthKey int

const (
	keyUserID MiddlewareAuthKey = iota // ключ для ID пользователя
)

// Сервис должен:
//   - Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя,
//     если такой куки не существует или она не проходит проверку подлинности.
//   - Если кука не содержит ID пользователя, хендлер должен возвращать HTTP-статус 401 Unauthorized.
func (c *ShortenerController) authJWTCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		ctx := r.Context()
		cookie, err := r.Cookie("Authorization")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				// если это запрос на получение URL'ов пользователя,
				// то возвращаем 401, т.к. куки не найден
				if r.URL.Path == "/api/user/urls" && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				userID, err = handleCookieError(ctx, w, c)
				if err != nil {
					logger.Log.Error("middleware auth error: ", zap.Error(err))
					http.Error(w, "unexpected auth error", http.StatusInternalServerError)
					return
				}
			} else {
				logger.Log.Error("middleware auth error: ", zap.Error(err))
				http.Error(w, "unexpected auth error", http.StatusInternalServerError)
				return
			}
		} else {
			var cookieValue string
			arr := strings.Split(cookie.Value, " ")
			if len(arr) == 2 {
				cookieValue = arr[1]
			}

			userID, err = common.ValidateAndGetUserID(cookieValue)
			// вернуть 401
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		ctx = context.WithValue(r.Context(), keyUserID, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// handleCookieError генерирует новый userID и записывает его в куки
func handleCookieError(ctx context.Context, w http.ResponseWriter, c *ShortenerController) (string, error) {
	userID, err := c.ucAuth.GetNewUserID(ctx)
	if err != nil {
		return "", err
	}
	err = createJWTAndSaveAsCookie(w, userID)
	if err != nil {
		return "", err
	}

	return userID, err
}

func createJWTAndSaveAsCookie(w http.ResponseWriter, userID string) error {
	jwt, err := common.BuildJWTString(userID)
	if err != nil {
		return err
	}
	bearerStr := fmt.Sprintf("Bearer %s", jwt)
	cookie := createAuthCookie(bearerStr)

	http.SetCookie(w, &cookie)
	return nil
}

func createAuthCookie(bearerStr string) http.Cookie {
	return http.Cookie{
		Name:     "Authorization",
		Value:    bearerStr,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
