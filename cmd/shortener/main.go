package main

import (
	"io"
	"math/rand"
	"net/http"

	"github.com/gorilla/mux"
)

func CompressURLHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}
	shortKey := generateShortKey()
	savedURLs[shortKey] = string(body)
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	io.WriteString(res, string("http://"+req.Host+"/"+shortKey))
}

func ShortURLByID(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	shortURLID := params["id"]
	if originalURL, ok := savedURLs[shortURLID]; ok {
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
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

var savedURLs = make(map[string]string)

func main() {
	router := mux.NewRouter()
	router.HandleFunc(`/`, CompressURLHandler).Methods("POST")
	router.HandleFunc(`/{id:\w+}`, ShortURLByID).Methods("GET")

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
