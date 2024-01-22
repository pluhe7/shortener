package app

import (
	"fmt"

	"github.com/pluhe7/shortener/internal/models"
	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

func (s *Server) ShortenURL(originalURL string) (string, error) {
	if len(originalURL) < 1 {
		return "", fmt.Errorf("url shouldn't be empty")
	}

	shortID := util.GetRandomString(idLen)

	err := s.Storage.Save(models.ShortURLRecord{
		ShortURL:    shortID,
		OriginalURL: originalURL})
	if err != nil {
		return "", fmt.Errorf("save to storage: %w", err)
	}

	return s.Config.BaseURL + "/" + shortID, nil
}

func (s *Server) ExpandURL(id string) (string, error) {
	if len([]rune(id)) != idLen {
		return "", fmt.Errorf("invalid url id")
	}

	expandedURL, err := s.Storage.Get(id)
	if err != nil {
		return "", err
	}

	return expandedURL, nil
}

func (s *Server) BatchShortenURLs(originalURLs []models.OriginalURLWithID) ([]models.ShortURLWithID, error) {
	records := make([]models.ShortURLRecord, 0, len(originalURLs))
	shortURLs := make([]models.ShortURLWithID, 0, len(originalURLs))

	for _, original := range originalURLs {
		shortID := util.GetRandomString(idLen)

		records = append(records, models.ShortURLRecord{
			ShortURL:    shortID,
			OriginalURL: original.OriginalURL})

		shortURLs = append(shortURLs, models.ShortURLWithID{
			CorrelationID: original.CorrelationID,
			ShortURL:      s.Config.BaseURL + "/" + shortID,
		})
	}

	err := s.Storage.SaveBatch(records)
	if err != nil {
		return nil, fmt.Errorf("save batch: %w", err)
	}

	return shortURLs, nil
}
