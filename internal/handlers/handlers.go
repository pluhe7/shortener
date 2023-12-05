package handlers

import (
	"github.com/gorilla/mux"
	"github.com/pluhe7/shortener/internal/app"
	"io"
	"net/http"
)

func ShortenHandler(w http.ResponseWriter, r *http.Request) {
	fullURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL := app.Shorten(string(fullURL))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)

	w.Write([]byte(shortURL))
}

func ExpandHandler(w http.ResponseWriter, r *http.Request) {
	expandedURL, err := app.Expand(mux.Vars(r)["id"])
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, expandedURL, http.StatusTemporaryRedirect)
}
