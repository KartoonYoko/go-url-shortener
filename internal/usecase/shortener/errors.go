package shortener

import "fmt"

type URLAlreadyExistsError struct {
	URL      string // URL, который уже существует в базе
	ShortURL string // короткий url
	Err      error
}

func NewURLAlreadyExistsError(id string, url string, err error) *URLAlreadyExistsError {
	return &URLAlreadyExistsError{
		URL:      url,
		ShortURL: id,
	}
}

func (e *URLAlreadyExistsError) Error() string {
	return fmt.Sprintf("url %s already exists", e.URL)
}

func (e *URLAlreadyExistsError) Unwrap() error {
	return e.Err
}
