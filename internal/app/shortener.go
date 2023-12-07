package app

import (
	"fmt"
	"github.com/pluhe7/shortener/internal/util"
	"net/url"
	"strings"
)

const idLen = 8

var shortURLs map[string]string

func ShortenURL(fullURL string) (string, error) {
	if len(fullURL) < 1 {
		return "", fmt.Errorf("url shouldn't be empty")
	}

	_, err := url.ParseRequestURI(fullURL)
	if err != nil {
		return "", fmt.Errorf("url validation error: %w", err)
	}

	shortID := util.GetRandomString(idLen)

	if shortURLs == nil {
		shortURLs = make(map[string]string)
	}

	shortURLs[shortID] = fullURL

	return "http://localhost:8080/" + shortID, nil
}

func ExpandURL(requestURLPath string) (string, error) {
	id := strings.TrimLeft(requestURLPath, "/")

	if len([]rune(id)) != idLen {
		return "", fmt.Errorf("invalid url id")
	}

	fullURL, ok := shortURLs[id]
	if !ok {
		return "", fmt.Errorf("url does not exist")
	}

	return fullURL, nil
}
