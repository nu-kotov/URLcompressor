package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/middleware"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc(`/`, middleware.RequestLogger(service.CompressURL))
	router.HandleFunc(`/{id:\w+}`, middleware.RequestLogger(service.RedirectByShortURLID))
	router.HandleFunc(`/api/shorten`, middleware.RequestLogger(service.GetShortURL))

	return router
}
