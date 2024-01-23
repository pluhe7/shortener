package storage

import (
	"context"

	"github.com/pluhe7/shortener/internal/models"
)

type MemoryStorage struct {
	ShortURLs map[string]string
}

func NewMemoryStorage() (*MemoryStorage, error) {
	storage := MemoryStorage{
		ShortURLs: make(map[string]string),
	}

	return &storage, nil
}

func (s *MemoryStorage) Get(shortURL string) (string, error) {
	originalURL, ok := s.ShortURLs[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	return originalURL, nil
}

func (s *MemoryStorage) GetByOriginal(originalURL string) (string, error) {
	var shortURL string

	for short, original := range s.ShortURLs {
		if original == originalURL {
			shortURL = short
			break
		}
	}

	if shortURL == "" {
		return "", ErrURLNotFound
	}

	return shortURL, nil
}

func (s *MemoryStorage) Save(record models.ShortURLRecord) error {
	s.ShortURLs[record.ShortURL] = record.OriginalURL

	return nil
}

func (s *MemoryStorage) SaveBatch(records []models.ShortURLRecord) error {
	for _, record := range records {
		s.ShortURLs[record.ShortURL] = record.OriginalURL
	}

	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) PingContext(ctx context.Context) error {
	return nil
}
