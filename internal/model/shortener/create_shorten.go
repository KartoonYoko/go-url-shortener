package shortener

// Запрос на создание сокращенного URL'a
type CreateShortenURLRequest struct {
	URL string `json:"url"`
}

// Ответ на запрос создание сокращенного URL'a
type CreateShortenURLResponse struct {
	Result string `json:"result"` // сокращенный URL
}

// Запрос на создание сокращенных URL'ов пачкой
type CreateShortenURLBatchItemRequest struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор
	OriginalURL   string `json:"original_url"`   // URL для сокращения
}

// Ответ на запрос создания сокращенных URL'ов пачкой
type CreateShortenURLBatchItemResponse struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор из объекта запроса
	ShortURL      string `json:"short_url"`      // результирующий сокращённый URL
}
