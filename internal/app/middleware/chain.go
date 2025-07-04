package middleware

import (
	"net/http"
)

type Middleware func(h http.HandlerFunc) http.HandlerFunc

// Chain возвращает миддлвары по цепочке, для последовательного вызова.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}
