package main

import (
	"io"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
)

func (uc *URLCompressor) postCompressURLHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}
	shortKey := generateShortKey()
	uc.urls[shortKey] = string(body)
	res.Header().Set("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, string(shortKey))
}

func (uc *URLCompressor) getShortURLById(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	shortUrlId := params["id"]
	originalURL, found := uc.urls[shortUrlId]
	if !found {
		http.Error(res, "Shortened key not found", http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusTemporaryRedirect)
	io.WriteString(res, originalURL)
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

	router := mux.NewRouter()
	router.HandleFunc(`/`, urlCompressor.postCompressURLHandler).Methods("POST")
	router.HandleFunc(`/{id:\w+}`, urlCompressor.getShortURLById).Methods("GET")

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
