package app

import (
	"fmt"

	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

var (
	ErrURLNotFound = fmt.Errorf("url does not exist")
)

func (s *Server) ShortenURL(fullURL string) (string, error) {
	if len(fullURL) < 1 {
		return "", fmt.Errorf("url shouldn't be empty")
	}

	shortID := util.GetRandomString(idLen)

	err := s.storage.Add(shortID, fullURL)
	if err != nil {
		return "", fmt.Errorf("add to storage: %w", err)
	}

	return s.Config.BaseURL + "/" + shortID, nil
}

func (s *Server) ExpandURL(id string) (string, error) {
	if len([]rune(id)) != idLen {
		return "", fmt.Errorf("invalid url id")
	}

	expandedURL, ok := s.storage.ShortURLs[id]
	if !ok {
		return "", ErrURLNotFound
	}

	return expandedURL, nil
}
