package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/handler"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(`/`, service.CompressURL)
	router.HandleFunc(`/{id:\w+}`, service.RedirectByShortURLID)

	return router
}
