package model

type CreateShortenURLRequest struct {
	URL string `json:"url"`
}

type CreateShortenURLResponse struct {
	Result string `json:"result"`
}
