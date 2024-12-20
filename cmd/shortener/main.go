package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct {
	config  Config
	storage map[string]string
}

func main() {
	var service Service
	service.config = ParseConfig()
	service.storage = make(map[string]string)

	router := mux.NewRouter()
	router.HandleFunc(`/`, service.CompressURLHandler)
	router.HandleFunc(`/{id:\w+}`, service.ShortURLByID)

	log.Fatal(http.ListenAndServe(service.config.RunAddr, router))
}
