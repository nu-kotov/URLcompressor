package models

// ShortenURLRequest - структура запроса, содержащая сокращенный урл.
type ShortenURLRequest struct {
	URL string `json:"url"`
}

// ShortenURLResponse - структура ответа, содержащая сокращенный урл.
type ShortenURLResponse struct {
	Result string `json:"result"`
}

// GetShortURLsBatchRequest - структура запроса, содержащая CorrelationID и сокращенный урл.
type GetShortURLsBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// GetShortURLsBatchResponse - структура ответа, содержащая CorrelationID и сокращенный урл.
type GetShortURLsBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// GetUserURLsResponse - структура ответа с сокращенным и полным урлом.
type GetUserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLsData - данные по урлу.
type URLsData struct {
	UserID        string `json:"user_id"`
	UUID          string `json:"uuid"`
	ShortURL      string `json:"short_url"`
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
	DeletedFlag   bool   `json:"is_deleted"`
}

// URLForDeleteMsg - структура сообщения для удаления урла.
type URLForDeleteMsg struct {
	UserID   string `json:"user_id"`
	ShortURL string `json:"short_url"`
}
