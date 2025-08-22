package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/service"
	"github.com/nu-kotov/URLcompressor/internal/app/auth"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"go.uber.org/zap"
)

// Handler - структура http хендлера для сокращения ссылок.
type Handler struct {
	Config        config.Config
	Storage       storage.Storage
	trustedSubnet *net.IPNet
	service       service.Service
}

// NewHandler - конструктор хендлера http сервиса для сокращения ссылок.
func NewHandler(config config.Config, service service.Service, storage storage.Storage, trustedSubnet *net.IPNet) *Handler {
	var hnd Handler

	hnd.Config = config
	hnd.Storage = storage
	hnd.trustedSubnet = trustedSubnet
	hnd.service = service

	return &hnd
}

// GetStats возвращает количество пользователей и урлов в сервисе.
func (hnd *Handler) GetStats(res http.ResponseWriter, req *http.Request) {
	if hnd.trustedSubnet == nil {
		http.Error(res, "trustedSubnet is nil", http.StatusForbidden)
		return
	}

	headerIPStr := req.Header.Get("X-Real-IP")
	if headerIPStr == "" {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	ip := net.ParseIP(headerIPStr)
	if ip == nil || !hnd.trustedSubnet.Contains(ip) {
		http.Error(res, "Not trusted IP", http.StatusForbidden)
		return
	}

	URLsCount, err := hnd.Storage.SelectURLsCount(req.Context())
	if err != nil {
		logger.Log.Info("Failed to get URLs count", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	usersCount, err := hnd.Storage.SelectUsersCount(req.Context())
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
func (hnd *Handler) DeleteUserURLs(res http.ResponseWriter, req *http.Request) {
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

	go hnd.service.SendURLsToDeletion(urls, userID)
}

// GetUserURLs возвращает список сокращенных и полных урлов пользователя.
func (hnd *Handler) GetUserURLs(res http.ResponseWriter, req *http.Request) {

	token, err := req.Cookie("token")

	if err != nil {
		res.WriteHeader(http.StatusUnauthorized)
	}

	userID, err := auth.GetUserID(token.Value)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
	}

	res.Header().Set("Content-Type", "application/json")
	if data, err := hnd.service.GetUserURLs(req.Context(), userID); data != nil {

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
func (hnd *Handler) GetShortURLsBatch(res http.ResponseWriter, req *http.Request) {
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

		response, err := hnd.service.GetShortURLsBatch(req.Context(), jsonBody, userID)
		if err != nil {
			logger.Log.Info(err.Error())
			http.Error(res, "Get short urls batch error", http.StatusInternalServerError)
			return
		}

		respJSON, err := json.Marshal(response)
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

// GetShortURL сохраняет сокращенный URL и возвращает его в качестве ответа.
func (hnd *Handler) GetShortURL(res http.ResponseWriter, req *http.Request) {

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

		resp, err := hnd.service.GetShortURLSrv(req.Context(), body, userID)

		if err != nil {
			if errors.Is(err, storage.ErrConflict) {

				respJSON, err := json.Marshal(resp)
				if err != nil {
					http.Error(res, err.Error(), http.StatusBadRequest)
					return
				}

				res.Header().Set("Content-Type", "application/json")
				res.WriteHeader(http.StatusConflict)
				res.Write(respJSON)
				return
			}
			logger.Log.Info(err.Error())
			http.Error(res, "Inserting to db error", http.StatusInternalServerError)
			return
		}

		respJSON, err := json.Marshal(resp)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
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
func (hnd *Handler) CompressURL(res http.ResponseWriter, req *http.Request) {

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

		shortID, err := hnd.service.CompressURL(req.Context(), body, userID)
		if err != nil {
			if errors.Is(err, storage.ErrConflict) {
				res.Header().Set("Content-Type", "text/plain")
				res.WriteHeader(http.StatusConflict)
				io.WriteString(res, string(hnd.Config.BaseURL+"/"+shortID))
				return
			}
			logger.Log.Info(err.Error())
			http.Error(res, "Inserting to db error", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		io.WriteString(res, string(hnd.Config.BaseURL+"/"+shortID))

	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

// RedirectByShortURLID редиректит по ID короткого урла на страницу по оригинальному урлу.
func (hnd *Handler) RedirectByShortURLID(res http.ResponseWriter, req *http.Request) {

	if req.Method == http.MethodGet {

		params := mux.Vars(req)
		shortURLID := params["id"]

		originalURL, err := hnd.service.SelectOriginalURLByShortURL(req.Context(), shortURLID)
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
func (hnd *Handler) PingDB(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		err := hnd.service.PingDB()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		} else {
			res.WriteHeader(http.StatusOK)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
