package main

import (
	"hash/fnv"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sqids/sqids-go"
)

func hash(bytes []byte) uint64 {

	h := fnv.New64a()
	h.Write([]byte(bytes))

	return h.Sum64()
}

func CompressURLHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		body, readingBodyError := io.ReadAll(req.Body)
		if readingBodyError != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}

		sqids, newSqidsError := sqids.New()
		if newSqidsError != nil {
			http.Error(res, "Sqids error", http.StatusBadRequest)
			return
		}

		bodyHash := hash(body)
		shortID, IDCreatingError := sqids.Encode([]uint64{bodyHash})
		if IDCreatingError != nil {
			http.Error(res, "IDCreatingError error", http.StatusBadRequest)
			return
		}

		savedURLs[shortID] = string(body)

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(FlagBaseURL+"/"+shortID))

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func ShortURLByID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		params := mux.Vars(req)
		shortURLID := params["id"]

		if originalURL, exists := savedURLs[shortURLID]; exists {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}

}

var savedURLs = make(map[string]string)

func main() {
	router := mux.NewRouter()
	router.HandleFunc(`/`, CompressURLHandler)
	router.HandleFunc(`/{id:\w+}`, ShortURLByID)

	ParseConfigFlags()
	listenAndServeError := http.ListenAndServe(FlagRunAddr, router)
	if listenAndServeError != nil {
		panic(listenAndServeError)
	}
}
