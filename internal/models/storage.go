package models

type ShortURLRecord struct {
	ID          int    `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"-"`
}
