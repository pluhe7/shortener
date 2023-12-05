package app

import (
	"fmt"
	"github.com/pluhe7/shortener/internal/util"
)

var shortURLs map[string]string

func Shorten(url string) string {
	shortID := util.GetRandomString(8)

	if shortURLs == nil {
		shortURLs = make(map[string]string)
	}

	shortURLs[shortID] = url

	return "http://localhost:8080/" + shortID
}

func Expand(shortID string) (string, error) {
	fullURL, ok := shortURLs[shortID]
	if !ok {
		return "", fmt.Errorf("url does not exist")
	}

	return fullURL, nil
}
