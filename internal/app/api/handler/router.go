package handler

import (
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/internal/app/middleware"
)

// NewRouter - конструктор роутера.
func NewRouter(handler Handler) *mux.Router {
	router := mux.NewRouter()

	middlewareStack := middleware.Chain(
		middleware.RequestCompressor,
		middleware.RequestLogger,
		middleware.RequestSession,
	)

	router.HandleFunc(`/ping`, handler.PingDB)
	router.HandleFunc(`/`, middlewareStack(handler.CompressURL))
	router.HandleFunc(`/api/shorten`, middlewareStack(handler.GetShortURL))
	router.HandleFunc(`/{id:\w+}`, middlewareStack(handler.RedirectByShortURLID))
	router.HandleFunc(`/api/shorten/batch`, middlewareStack(handler.GetShortURLsBatch))
	router.HandleFunc(`/api/user/urls`, middlewareStack(handler.GetUserURLs)).Methods("GET")
	router.HandleFunc(`/api/user/urls`, middlewareStack(handler.DeleteUserURLs)).Methods("DELETE")
	router.HandleFunc(`/api/internal/stats`, middlewareStack(handler.DeleteUserURLs)).Methods("GET")

	return router
}
