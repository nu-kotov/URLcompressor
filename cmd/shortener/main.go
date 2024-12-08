package main

import (
	"io"
	"math/rand"
	"net/http"
)

func (uc *URLCompressor) compressURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}
		shortKey := generateShortKey()
		uc.urls[shortKey] = string(body)
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(shortKey))

	} else if req.Method == http.MethodGet {

		shortURLID := req.URL.Path[1:]
		if originalURL, ok := uc.urls[shortURLID]; ok {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}

}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}

type URLCompressor struct {
	urls map[string]string
}

func main() {
	urlCompressor := &URLCompressor{
		urls: make(map[string]string),
	}

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, urlCompressor.compressURLHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
