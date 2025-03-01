package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/middleware"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()

	middlewareStack := middleware.Chain(
		middleware.RequestCompressor,
		middleware.RequestLogger,
	)

	router.HandleFunc(`/ping`, service.PingDB)
	router.HandleFunc(`/`, middlewareStack(service.CompressURL))
	router.HandleFunc(`/api/shorten`, middlewareStack(service.GetShortURL))
	router.HandleFunc(`/{id:\w+}`, middlewareStack(service.RedirectByShortURLID))
	router.HandleFunc(`/api/shorten/batch`, middlewareStack(service.GetShortURLsBatch))

	return router
}
