package app

import (
	"errors"
	"fmt"

	"github.com/pluhe7/shortener/internal/models"
	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

var ErrEmptyURL = errors.New("url shouldn't be empty")

func (s *Server) ShortenURL(originalURL, userID string) (string, error) {
	if len(originalURL) < 1 {
		return "", ErrEmptyURL
	}

	shortID := util.GetRandomString(idLen)
	shortURL := s.getShortURLFromID(shortID)

	err := s.Storage.Save(models.ShortURLRecord{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
	})
	if err != nil {
		return "", fmt.Errorf("save to storage: %w", err)
	}

	return shortURL, nil
}

func (s *Server) ExpandURL(id string) (*models.ShortURLRecord, error) {
	if len([]rune(id)) != idLen {
		return nil, errors.New("invalid url id")
	}

	record, err := s.Storage.Get(s.getShortURLFromID(id))
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (s *Server) BatchShortenURLs(originalURLs []models.OriginalURLWithID, userID string) ([]models.ShortURLWithID, error) {
	records := make([]models.ShortURLRecord, 0, len(originalURLs))
	shortURLs := make([]models.ShortURLWithID, 0, len(originalURLs))

	for _, original := range originalURLs {
		shortID := util.GetRandomString(idLen)
		shortURL := s.getShortURLFromID(shortID)

		records = append(records, models.ShortURLRecord{
			ShortURL:    shortURL,
			OriginalURL: original.OriginalURL,
			UserID:      userID,
		})

		shortURLs = append(shortURLs, models.ShortURLWithID{
			CorrelationID: original.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	err := s.Storage.SaveBatch(records)
	if err != nil {
		return nil, fmt.Errorf("save batch: %w", err)
	}

	return shortURLs, nil
}

func (s *Server) GetExistingShortURL(originalURL string) (string, error) {
	record, err := s.Storage.GetByOriginal(originalURL)
	if err != nil {
		return "", fmt.Errorf("get by original: %w", err)
	}

	return record.ShortURL, nil
}

func (s *Server) getShortURLFromID(id string) string {
	return s.Config.BaseURL + "/" + id
}
