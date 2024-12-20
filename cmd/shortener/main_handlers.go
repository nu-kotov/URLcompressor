package main

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sqids/sqids-go"
)

func (srv Service) CompressURLHandler(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		body, errReadingBody := io.ReadAll(req.Body)
		if errReadingBody != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}

		sqids, errNewSqids := sqids.New()
		if errNewSqids != nil {
			http.Error(res, "Sqids lib error", http.StatusInternalServerError)
			return
		}

		bodyHash := hash(body)
		shortID, errIDCreating := sqids.Encode([]uint64{bodyHash})
		if errIDCreating != nil {
			http.Error(res, "Short ID creating error", http.StatusInternalServerError)
			return
		}

		srv.storage[shortID] = string(body)

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(srv.config.BaseURL+"/"+shortID))

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv Service) ShortURLByID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		params := mux.Vars(req)
		shortURLID := params["id"]

		if originalURL, exists := srv.storage[shortURLID]; exists {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
