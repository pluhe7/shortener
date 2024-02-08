package storage

import (
	"context"
	"errors"

	"github.com/pluhe7/shortener/internal/models"
)

type MemoryStorage struct {
	ShortURLs map[string]models.ShortURLRecord
}

func NewMemoryStorage() (*MemoryStorage, error) {
	storage := MemoryStorage{
		ShortURLs: make(map[string]models.ShortURLRecord),
	}

	return &storage, nil
}

func (s *MemoryStorage) Get(shortURL string) (*models.ShortURLRecord, error) {
	record, ok := s.ShortURLs[shortURL]
	if !ok {
		return nil, ErrURLNotFound
	}

	return &record, nil
}

func (s *MemoryStorage) GetByOriginal(originalURL string) (*models.ShortURLRecord, error) {
	for _, record := range s.ShortURLs {
		if record.OriginalURL == originalURL {
			return &record, nil
		}
	}

	return nil, ErrURLNotFound
}

func (s *MemoryStorage) FindByUserID(userID string) ([]models.ShortURLRecord, error) {
	var records []models.ShortURLRecord

	for _, record := range s.ShortURLs {
		if record.UserID == userID {
			records = append(records, record)
		}
	}

	return records, nil
}

func (s *MemoryStorage) Save(record models.ShortURLRecord) error {
	s.ShortURLs[record.ShortURL] = record

	return nil
}

func (s *MemoryStorage) SaveBatch(records []models.ShortURLRecord) error {
	for _, record := range records {
		s.ShortURLs[record.ShortURL] = record
	}

	return nil
}

func (s *MemoryStorage) Delete(shortURLs []string) error {
	for _, shortURL := range shortURLs {
		if record, ok := s.ShortURLs[shortURL]; ok {
			record.IsDeleted = true

			s.ShortURLs[shortURL] = record

		} else {
			return errors.New("record doesn't exist")
		}
	}

	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) PingContext(ctx context.Context) error {
	return nil
}
