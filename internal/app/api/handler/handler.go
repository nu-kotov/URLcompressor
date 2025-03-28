package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"fmt"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/utils"
	"github.com/nu-kotov/URLcompressor/internal/app/auth"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"github.com/sqids/sqids-go"
)

type Service struct {
	Config  config.Config
	Storage storage.Storage
}

func NewService(config config.Config, storage storage.Storage) *Service {
	var srv Service

	srv.Config = config
	srv.Storage = storage

	return &srv
}

func (srv *Service) GetUserURLs(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {

		fmt.Println("Вызывается GetUserURLs")

		token, err := req.Cookie("token")

		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
		}

		userID, err := auth.GetUserID(token.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		fmt.Println("userID!!!", userID)
		res.Header().Set("Content-Type", "application/json")
		if data, err := srv.Storage.SelectURLs(req.Context(), userID); data != nil {

			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}

			resp, err := json.Marshal(data)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}

			_, err = res.Write(resp)

			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
			}

		} else {
			res.WriteHeader(http.StatusNoContent)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) GetShortURLsBatch(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

		fmt.Println("Вызывается GetShortURLsBatch")

		token, err := req.Cookie("token")

		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
		}

		userID, err := auth.GetUserID(token.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

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
					ShortURL:      srv.Config.BaseURL + "/" + shortID,
				},
			)

			strBody := string(row.OriginalURL)

			event := models.URLsData{
				UUID:          uuid.New().String(),
				ShortURL:      shortID,
				OriginalURL:   strBody,
				CorrelationID: row.CorrelationID,
				UserID:        userID,
			}
			rowsBatch = append(rowsBatch, event)
		}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		err = srv.Storage.InsertURLsDataBatch(req.Context(), rowsBatch)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, "Inserting to db error", http.StatusInternalServerError)
			return
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

		token, err := req.Cookie("token")

		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
		}

		userID, err := auth.GetUserID(token.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

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

		resp := models.ShortenURLResponse{Result: srv.Config.BaseURL + "/" + shortID}
		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

		strBody := string(jsonBody.URL)
		event := models.URLsData{UserID: userID, UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody}

		err = srv.Storage.InsertURLsData(req.Context(), &event)

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

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(respJSON)

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) CompressURL(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

		fmt.Println("Вызывается CompressURL")

		token, err := req.Cookie("token")

		if err != nil {
			res.WriteHeader(http.StatusUnauthorized)
		}

		userID, err := auth.GetUserID(token.Value)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
		}

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
		event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody, UserID: userID}

		err = srv.Storage.InsertURLsData(req.Context(), &event)
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				res.Header().Set("Content-Type", "text/plain")
				res.WriteHeader(http.StatusConflict)
				io.WriteString(res, string(srv.Config.BaseURL+"/"+shortID))
				return
			}
			logger.Log.Info(err.Error())
			http.Error(res, "Inserting to db error", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(srv.Config.BaseURL+"/"+shortID))

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) RedirectByShortURLID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		fmt.Println("Вызывается RedirectByShortURLID")

		params := mux.Vars(req)
		shortURLID := params["id"]

		originalURL, err := srv.Storage.SelectOriginalURLByShortURL(req.Context(), shortURLID)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, "URLs select error", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func (srv *Service) PingDB(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		err := srv.Storage.Ping()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		} else {
			res.WriteHeader(http.StatusOK)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
