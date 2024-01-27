package shortener

import (
	"errors"
	"fmt"
)

var ErrNotFoundKey = errors.New("service: key not found")

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
