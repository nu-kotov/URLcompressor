package main

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/internal/app/api/handler"
	"github.com/nu-kotov/URLcompressor/internal/app/middleware"
)

func NewRouter(service handler.Service) *mux.Router {
	router := mux.NewRouter()

	middlewareStack := middleware.Chain(
		middleware.RequestCompressor,
		middleware.RequestLogger,
		middleware.RequestSession,
	)

	router.HandleFunc(`/ping`, service.PingDB)
	router.HandleFunc(`/`, middlewareStack(service.CompressURL))
	router.HandleFunc(`/api/shorten`, middlewareStack(service.GetShortURL))
	router.HandleFunc(`/{id:\w+}`, middlewareStack(service.RedirectByShortURLID))
	router.HandleFunc(`/api/shorten/batch`, middlewareStack(service.GetShortURLsBatch))
	router.HandleFunc(`/api/user/urls`, middlewareStack(service.GetUserURLs))

	return router
}
