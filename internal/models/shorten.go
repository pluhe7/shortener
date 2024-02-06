package models

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type OriginalURLWithID struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortURLWithID struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
