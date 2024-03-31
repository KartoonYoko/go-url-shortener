package repository

import (
	"errors"
	"fmt"
)

// Ошибки, которые могут возникнуть при использовании хранилища
var (
	ErrNotFoundKey = errors.New("repository: key not found")
	ErrURLDeleted  = errors.New("repository: url was removed")
)

// URLAlreadyExistsError говорит о том, что переданный URL уже существует в БД
type URLAlreadyExistsError struct {
	URL string // URL, который уже существует в базе
	ID  string // ID url'a
}

// Конструктор для URLAlreadyExistsError
func NewURLAlreadyExistsError(id string, url string) *URLAlreadyExistsError {
	return &URLAlreadyExistsError{
		URL: url,
		ID:  id,
	}
}

func (e *URLAlreadyExistsError) Error() string {
	return fmt.Sprintf("url %s already exists", e.URL)
}
