package app

import (
	"fmt"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

var (
	ErrEmptyURL    = fmt.Errorf("url shouldn't be empty")
	ErrInvalidURL  = fmt.Errorf("invalid url id")
	ErrURLNotFound = fmt.Errorf("url does not exist")
)

var shortURLs map[string]string

func ShortenURL(fullURL string) (string, error) {
	if len(fullURL) < 1 {
		return "", ErrEmptyURL
	}

	shortID := util.GetRandomString(idLen)

	if shortURLs == nil {
		shortURLs = make(map[string]string)
	}

	shortURLs[shortID] = fullURL

	return config.GetConfig().BaseURL + "/" + shortID, nil
}

func ExpandURL(id string) (string, error) {
	if len([]rune(id)) != idLen {
		return "", ErrInvalidURL
	}

	expandedURL, ok := shortURLs[id]
	if !ok {
		return "", ErrURLNotFound
	}

	return expandedURL, nil
}
