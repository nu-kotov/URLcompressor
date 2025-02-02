package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/utils"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"github.com/sqids/sqids-go"
)

type Service struct {
	config      config.Config
	mapStorage  map[string]string
	fileStorage storage.FileStorage
}

func InitService(config config.Config) (Service, error) {
	var srv Service
	srv.config = config

	srv.fileStorage.DataConsumer, _ = storage.NewConsumer(config.FileStoragePath)
	srv.fileStorage.DataProducer, _ = storage.NewProducer(config.FileStoragePath)

	var err error
	srv.mapStorage, err = srv.fileStorage.DataConsumer.InitMapStorage()
	if err != nil {
		return srv, err
	}

	return srv, nil
}

func (srv *Service) GetShortURL(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		var jsonBody models.ShortenURLRequest

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(body, &jsonBody); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		sqids, err := sqids.New()
		if err != nil {
			http.Error(res, "Sqids lib error", http.StatusInternalServerError)
			return
		}

		bodyHash := utils.Hash([]byte(jsonBody.URL))
		shortID, err := sqids.Encode([]uint64{bodyHash})
		if err != nil {
			http.Error(res, "Short ID creating error", http.StatusInternalServerError)
			return
		}
		srv.mapStorage[shortID] = string(jsonBody.URL)

		resp := models.ShortenURLResponse{Result: srv.config.BaseURL + "/" + shortID}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		if _, exist := srv.mapStorage[shortID]; !exist {
			strBody := string(body)
			srv.mapStorage[shortID] = strBody

			event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody}
			srv.fileStorage.DataProducer.WriteEvent(&event)
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(respJSON)

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) CompressURL(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}

		sqids, err := sqids.New()
		if err != nil {
			http.Error(res, "Sqids lib error", http.StatusInternalServerError)
			return
		}

		bodyHash := utils.Hash(body)
		shortID, err := sqids.Encode([]uint64{bodyHash})
		if err != nil {
			http.Error(res, "Short ID creating error", http.StatusInternalServerError)
			return
		}

		if _, exist := srv.mapStorage[shortID]; !exist {
			strBody := string(body)
			srv.mapStorage[shortID] = strBody

			event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody}
			srv.fileStorage.DataProducer.WriteEvent(&event)
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(srv.config.BaseURL+"/"+shortID))

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) RedirectByShortURLID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		params := mux.Vars(req)
		shortURLID := params["id"]

		if originalURL, exists := srv.mapStorage[shortURLID]; exists {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
