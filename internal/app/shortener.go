package app

import (
	"fmt"

	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

func (s *Server) ShortenURL(originalURL string) (string, error) {
	if len(originalURL) < 1 {
		return "", fmt.Errorf("url shouldn't be empty")
	}

	shortID := util.GetRandomString(idLen)

	err := s.Storage.Add(shortID, originalURL)
	if err != nil {
		return "", fmt.Errorf("add to storage: %w", err)
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
