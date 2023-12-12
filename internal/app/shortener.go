package app

import (
	"fmt"
	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/util"
)

const idLen = 8

var shortURLs map[string]string

func ShortenURL(fullURL string) (string, error) {
	if len(fullURL) < 1 {
		return "", fmt.Errorf("url shouldn't be empty")
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
		return "", fmt.Errorf("invalid url id")
	}

	expandedURL, ok := shortURLs[id]
	if !ok {
		return "", fmt.Errorf("url does not exist")
	}

	return expandedURL, nil
}
