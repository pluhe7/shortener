package storage

import (
	"context"
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

func (s *MemoryStorage) Add(shortURL, originalURL string) error {
	s.ShortURLs[shortURL] = originalURL

	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) PingContext(ctx context.Context) error {
	return nil
}
