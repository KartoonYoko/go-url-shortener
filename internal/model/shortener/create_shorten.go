package shortener

// CreateShortenURLRequest Запрос на создание сокращенного URL'a
type CreateShortenURLRequest struct {
	URL string `json:"url"`
}

// CreateShortenURLResponse Ответ на запрос создание сокращенного URL'a
type CreateShortenURLResponse struct {
	Result string `json:"result"` // сокращенный URL
}

// CreateShortenURLBatchItemRequest Запрос на создание сокращенных URL'ов пачкой
type CreateShortenURLBatchItemRequest struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор
	OriginalURL   string `json:"original_url"`   // URL для сокращения
}

// CreateShortenURLBatchItemResponse Ответ на запрос создания сокращенных URL'ов пачкой
type CreateShortenURLBatchItemResponse struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор из объекта запроса
	ShortURL      string `json:"short_url"`      // результирующий сокращённый URL
}
