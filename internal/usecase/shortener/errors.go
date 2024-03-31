package shortener

import (
	"errors"
	"fmt"
)

// ErrURLDeleted сообщает, что URL удалён
var ErrURLDeleted = errors.New("service: url was removed")

// URLAlreadyExistsError сигнализирует, что URL уже существует
type URLAlreadyExistsError struct {
	URL      string // URL, который уже существует в базе
	ShortURL string // короткий url
	Err      error
}

// NewURLAlreadyExistsError конструктор
func NewURLAlreadyExistsError(id string, url string, err error) *URLAlreadyExistsError {
	return &URLAlreadyExistsError{
		URL:      url,
		ShortURL: id,
	}
}

// Error реализует error
func (e *URLAlreadyExistsError) Error() string {
	return fmt.Sprintf("url %s already exists", e.URL)
}

// Unwrap
func (e *URLAlreadyExistsError) Unwrap() error {
	return e.Err
}
