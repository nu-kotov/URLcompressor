package models

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	Result string `json:"result"`
}

type GetShortURLsBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type GetShortURLsBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type GetUserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLsData struct {
	UserID        string `json:"user_id"`
	UUID          string `json:"uuid"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
	DeletedFlag   bool   `json:"is_deleted"`
}
