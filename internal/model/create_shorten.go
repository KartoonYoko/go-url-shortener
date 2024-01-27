package model

type CreateShortenURLRequest struct {
	URL string `json:"url"`
}

type CreateShortenURLResponse struct {
	Result string `json:"result"`
}

type CreateShortenURLBatchItemRequest struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор
	OriginalURL   string `json:"original_url"`   // URL для сокращения
}

type CreateShortenURLBatchItemResponse struct {
	CorrelationID string `json:"correlation_id"` // строковый идентификатор из объекта запроса
	ShortURL      string `json:"short_url"`      // результирующий сокращённый URL
}
