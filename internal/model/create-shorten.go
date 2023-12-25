package model

type CreateShortenURLRequest struct {
	Url string `json:"url"`
}

type CreateShortenURLResponse struct {
	Result string `json:"result"`
}