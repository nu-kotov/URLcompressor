package handler

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/internal/app/middleware"
)

// NewRouter - конструктор роутера.
func NewRouter(service Service) *mux.Router {
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
	router.HandleFunc(`/api/user/urls`, middlewareStack(service.GetUserURLs)).Methods("GET")
	router.HandleFunc(`/api/user/urls`, middlewareStack(service.DeleteUserURLs)).Methods("DELETE")
	router.HandleFunc(`/api/internal/stats`, middlewareStack(service.DeleteUserURLs)).Methods("GET")

	return router
}
