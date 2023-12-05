package main

import (
	"github.com/gorilla/mux"
	"github.com/pluhe7/shortener/internal/handlers"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc(`/`, handlers.ShortenHandler)
	router.HandleFunc(`/{id}`, handlers.ExpandHandler)

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
