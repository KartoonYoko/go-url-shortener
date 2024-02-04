package repository

import (
	"errors"
	"fmt"
)

var ErrNotFoundKey = errors.New("repository: key not found")
var ErrURLDeleted = errors.New("repository: url was removed")

type URLAlreadyExistsError struct {
	URL string // URL, который уже существует в базе
	ID  string // ID url'a
}

func NewURLAlreadyExistsError(id string, url string) *URLAlreadyExistsError {
	return &URLAlreadyExistsError{
		URL: url,
		ID:  id,
	}
}

func (e *URLAlreadyExistsError) Error() string {
	return fmt.Sprintf("url %s already exists", e.URL)
}
