package shortener

// URL'ы пользователя
type GetUserURLsItemResponse struct {
	ShortURL    string `json:"short_url"`    // сокращённый URL
	OriginalURL string `json:"original_url"` // оригинальный URL
}
