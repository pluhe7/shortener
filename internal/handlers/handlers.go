package handlers

import (
	"github.com/pluhe7/shortener/internal/app"
	"io"
	"net/http"
)

func BaseHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		ExpandHandler(w, r)
		return

	case http.MethodPost:
		ShortenHandler(w, r)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong method"))
		return
	}
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	expandedURL, err := app.ExpandURL(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, expandedURL, http.StatusTemporaryRedirect)
}

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	fullURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("wrong request body: " + err.Error()))
		return
	}

	shortURL, err := app.ShortenURL(string(fullURL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("shorten url error: " + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortURL))
}
