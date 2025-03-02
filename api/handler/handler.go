package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/api/utils"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"github.com/sqids/sqids-go"
)

type Service struct {
	config      config.Config
	mapStorage  map[string]string
	fileStorage *storage.FileStorage
	DBStorage   *storage.DBStorage
}

func InitService(config config.Config) (*Service, error) {
	var srv Service
	var err error

	srv.config = config
	srv.mapStorage = make(map[string]string)

	if config.FileStoragePath != "" {
		srv.fileStorage, err = storage.NewFileStorage(config.FileStoragePath)
		if err != nil {
			return nil, err
		}

		srv.mapStorage, err = srv.fileStorage.InitMapStorage()
		if err != nil {
			return nil, err
		}
	}

	if config.DatabaseConnection != "" {
		srv.DBStorage, err = storage.NewConnect(config.DatabaseConnection)
		if err != nil {
			return nil, err
		}

		err = srv.DBStorage.CreateTable()
		if err != nil {
			return nil, err
		}
	}

	return &srv, nil
}

func (srv *Service) GetShortURLsBatch(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		var jsonBody []models.GetShortURLsBatchRequest

		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Invalid body", http.StatusBadRequest)
			return
		}
		if err = json.Unmarshal(body, &jsonBody); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		var resp []models.GetShortURLsBatchResponse
		var rowsBatch []models.URLsData
		for _, row := range jsonBody {

			sqids, err := sqids.New()
			if err != nil {
				http.Error(res, "Sqids lib error", http.StatusInternalServerError)
				return
			}

			bodyHash := utils.Hash([]byte(row.OriginalURL))
			shortID, err := sqids.Encode([]uint64{bodyHash})
			if err != nil {
				http.Error(res, "Short ID creating error", http.StatusInternalServerError)
				return
			}

			resp = append(
				resp,
				models.GetShortURLsBatchResponse{
					CorrelationID: row.CorrelationID,
					ShortURL:      srv.config.BaseURL + "/" + shortID,
				},
			)

			if _, exist := srv.mapStorage[shortID]; !exist {

				strBody := string(row.OriginalURL)
				srv.mapStorage[shortID] = strBody

				event := models.URLsData{
					UUID:          uuid.New().String(),
					ShortURL:      shortID,
					OriginalURL:   strBody,
					CorrelationID: row.CorrelationID,
				}

				if srv.config.FileStoragePath != "" {
					srv.fileStorage.ProduceEvent(&event)
				}
				rowsBatch = append(rowsBatch, event)
			}
		}
		if srv.config.DatabaseConnection != "" {
			err := srv.DBStorage.InsertURLsDataBatch(req.Context(), rowsBatch)
			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(res, "Inserting to db error", http.StatusInternalServerError)
				return
			}
		}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(respJSON)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
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

		resp := models.ShortenURLResponse{Result: srv.config.BaseURL + "/" + shortID}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		strBody := string(jsonBody.URL)
		event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody}

		if _, exist := srv.mapStorage[shortID]; !exist {

			srv.mapStorage[shortID] = strBody

			if srv.config.FileStoragePath != "" {
				srv.fileStorage.ProduceEvent(&event)
			}
		}

		if srv.config.DatabaseConnection != "" {
			err := srv.DBStorage.InsertURLsData(req.Context(), &event)
			if err != nil {
				if errors.Is(err, storage.ErrConflict) {
					res.Header().Set("Content-Type", "application/json")
					res.WriteHeader(http.StatusConflict)
					res.Write(respJSON)
					return
				}
				logger.Log.Info(err.Error())
				http.Error(res, "Inserting to db error", http.StatusInternalServerError)
				return
			}
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

		strBody := string(body)
		event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody}

		if _, exist := srv.mapStorage[shortID]; !exist {

			srv.mapStorage[shortID] = strBody

			if srv.config.FileStoragePath != "" {
				srv.fileStorage.ProduceEvent(&event)
			}
		}

		if srv.config.DatabaseConnection != "" {
			err := srv.DBStorage.InsertURLsData(req.Context(), &event)
			if err != nil {
				if errors.Is(err, storage.ErrConflict) {
					res.Header().Set("Content-Type", "text/plain")
					res.WriteHeader(http.StatusConflict)
					io.WriteString(res, string(srv.config.BaseURL+"/"+shortID))
					return
				}
				logger.Log.Info(err.Error())
				http.Error(res, "Inserting to db error", http.StatusInternalServerError)
				return
			}
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

		if srv.config.DatabaseConnection != "" {
			originalURL, err := srv.DBStorage.SelectOriginalURLByShortURL(req.Context(), shortURLID)
			if err != nil {
				logger.Log.Info(err.Error())
				http.Error(res, "URLs select error", http.StatusInternalServerError)
				return
			}
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else if originalURL, exists := srv.mapStorage[shortURLID]; exists {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) PingDB(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		if srv.config.DatabaseConnection == "" {
			res.WriteHeader(http.StatusInternalServerError)
		} else {
			err := srv.DBStorage.Ping()
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
			} else {
				res.WriteHeader(http.StatusOK)
			}
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
