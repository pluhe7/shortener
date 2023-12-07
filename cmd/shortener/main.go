package main

import (
	"github.com/pluhe7/shortener/internal/handlers"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc(`/`, handlers.BaseHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
