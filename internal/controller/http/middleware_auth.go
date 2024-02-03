package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Claims — структура утверждений, которая включает стандартные утверждения
// и одно пользовательское — UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// тип ключа контекста для middleware аутентификации
type MiddlewareAuthKey int

const (
	keyUserID MiddlewareAuthKey = iota // ключ для ID пользователя
)

// const TOKEN_EXP = time.Hour * 3
const SecretKey = "supersecretkey"

// Сервис должен:
//   - Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя,
//     если такой куки не существует или она не проходит проверку подлинности.
//   - Если кука не содержит ID пользователя, хендлер должен возвращать HTTP-статус 401 Unauthorized.
func (c *shortenerController) authJWTCookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// - получить куку
		// - если существует
		// 		- провалидировать JWT из куки
		// 			- если не содержит ID, вернуть 401
		// 			- если не валиден, то собрать JWT вместе с нвоым UserID и добавить в куки
		// - если не существует
		// 		- собрать JWT вместе с нвоым UserID и добавить в куки

		var userID string
		ctx := r.Context()
		cookie, err := r.Cookie("Authorization")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
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

			userID, err = validateAndGetUserID(cookieValue)
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
func handleCookieError(ctx context.Context, w http.ResponseWriter, c *shortenerController) (string, error) {
	userID, err := c.ucAuth.GetNewUserID(ctx)
	if err != nil {
		return "", err
	}
	err = createJWTAndSaveAsCookieAnHeader(w, userID)
	if err != nil {
		return "", err
	}

	return userID, err
}

func createJWTAndSaveAsCookieAnHeader(w http.ResponseWriter, userID string) error {
	jwt, err := buildJWTString(userID)
	if err != nil {
		return err
	}
	bearerStr := fmt.Sprintf("Bearer %s", jwt)

	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    bearerStr,
		Path:     "/",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.Header().Add("Authorization", bearerStr)
	return nil
}

// buildJWTString создаёт токен и возвращает его в виде строки.
func buildJWTString(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// validateAndGetUserID валидирует токен и получает из него UserID
func validateAndGetUserID(tokenString string) (string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SecretKey), nil
		})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("token is not valid")
	}

	return claims.UserID, nil
}
