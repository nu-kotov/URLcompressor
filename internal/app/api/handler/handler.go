package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/utils"
	"github.com/nu-kotov/URLcompressor/internal/app/auth"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"github.com/sqids/sqids-go"
	"go.uber.org/zap"
)

// Service - структура сервиса для сокращения ссылок.
type Service struct {
	Config         config.Config
	Storage        storage.Storage
	URLsDeletionCh chan models.URLForDeleteMsg
	trustedSubnet  *net.IPNet
}

// NewService - конструктор сервиса для сокращения ссылок.
func NewService(config config.Config, storage storage.Storage, trustedSubnet *net.IPNet) *Service {
	var srv Service

	srv.Config = config
	srv.Storage = storage
	srv.trustedSubnet = trustedSubnet
	srv.URLsDeletionCh = make(chan models.URLForDeleteMsg, 1024)

	go srv.flushMessages()

	return &srv
}

// GetStats возвращает количество пользователей и урлов в сервисе.
func (srv *Service) GetStats(res http.ResponseWriter, req *http.Request) {
	if srv.trustedSubnet == nil {
		http.Error(res, "trustedSubnet is nil", http.StatusForbidden)
		return
	}

	headerIPStr := req.Header.Get("X-Real-IP")
	if headerIPStr == "" {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	ip := net.ParseIP(headerIPStr)
	if ip == nil || !srv.trustedSubnet.Contains(ip) {
		http.Error(res, "Not trusted IP", http.StatusForbidden)
		return
	}

	URLsCount, err := srv.Storage.SelectURLsCount(req.Context())
	if err != nil {
		logger.Log.Info("Failed to get URLs count", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	usersCount, err := srv.Storage.SelectUsersCount(req.Context())
	if err != nil {
		logger.Log.Info("Failed to get users count", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)

	resp := models.GetStatsResponse{URLs: URLsCount, Users: usersCount}
	JSONResp, err := json.Marshal(resp)
	if err != nil {
		logger.Log.Info("Failed to marshal response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = res.Write(JSONResp)

	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteUserURLs помещает список ID урлов ["IhqFu4fdBD9w", "50ZT5FOYE6y"] в канал для удаления.
func (srv *Service) DeleteUserURLs(res http.ResponseWriter, req *http.Request) {
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
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var urls []string
	if err = json.Unmarshal(body, &urls); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusAccepted)

	for _, url := range urls {
		srv.URLsDeletionCh <- models.URLForDeleteMsg{
			ShortURL: url,
			UserID:   userID,
		}
	}
}

// GetUserURLs возвращает список сокращенных и полных урлов пользователя.
func (srv *Service) GetUserURLs(res http.ResponseWriter, req *http.Request) {

	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

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
}

// GetShortURLsBatch сохраняет батч коротких урлов и возвращает его в качестве ответа.
func (srv *Service) GetShortURLsBatch(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {

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

// GetShortURL сохраняет сокращенный URL и возвращает его в качестве ответа.
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

// CompressURL сохраняет сокращенный URL и возвращает его в качестве ответа.
func (srv *Service) CompressURL(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodPost {

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

// RedirectByShortURLID редиректит по ID короткого урла на страницу по оригинальному урлу.
func (srv *Service) RedirectByShortURLID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		params := mux.Vars(req)
		shortURLID := params["id"]

		originalURL, err := srv.Storage.SelectOriginalURLByShortURL(req.Context(), shortURLID)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, "URLs select error", http.StatusInternalServerError)
			return
		}

		if originalURL == "deleted" {
			res.WriteHeader(http.StatusGone)
			return
		}

		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

// PingDB пингует бд.
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

// flushMessages слушает сообщения в канале URLsDeletionCh, помечает, что урлы из сообщений удалены.
func (srv *Service) flushMessages() {
	ticker := time.NewTicker(10 * time.Second)

	var URLsForDelete []models.URLForDeleteMsg

	for {
		select {

		case msg := <-srv.URLsDeletionCh:
			URLsForDelete = append(URLsForDelete, msg)

		case <-ticker.C:
			if len(URLsForDelete) == 0 {
				continue
			}

			err := srv.Storage.DeleteURLs(context.Background(), URLsForDelete)
			if err != nil {
				logger.Log.Info(err.Error())
				continue
			}

			URLsForDelete = nil
		}
	}
}
