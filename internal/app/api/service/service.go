package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/api/utils"
	"github.com/nu-kotov/URLcompressor/internal/app/logger"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
	"github.com/nu-kotov/URLcompressor/internal/app/storage"
	"go.uber.org/zap"
)

// Service - интерфейс для работы с URL.
type Service interface {
	CompressURL(context.Context, []byte, string) (string, error)
	GetUserURLs(context.Context, string) ([]models.GetUserURLsResponse, error)
	GetShortURLsBatch(context.Context, []models.GetShortURLsBatchRequest, string) ([]models.GetShortURLsBatchResponse, error)
	GetShortURLSrv(context.Context, []byte, string) (*models.ShortenURLResponse, error)
	SendURLsToDeletion([]string, string)
	SelectOriginalURLByShortURL(context.Context, string) (string, error)
	GetStats(context.Context) (*models.GetStatsResponse, error)
	PingDB() error
}

// URLService - структура сервиса для сокращения ссылок.
type URLService struct {
	Config         config.Config
	Storage        storage.Storage
	URLsDeletionCh chan models.URLForDeleteMsg
}

// NewURLService - конструктор сервиса для сокращения ссылок.
func NewURLService(config config.Config, storage storage.Storage) *URLService {
	var srv URLService

	srv.Config = config
	srv.Storage = storage
	srv.URLsDeletionCh = make(chan models.URLForDeleteMsg, 1024)

	go srv.flushMessages()

	return &srv
}

// GetShortURLsBatch сохраняет батч коротких урлов и возвращает его в качестве ответа.
func (srv *URLService) GetShortURLsBatch(ctx context.Context, shortURLsBatch []models.GetShortURLsBatchRequest, userID string) ([]models.GetShortURLsBatchResponse, error) {
	var resp []models.GetShortURLsBatchResponse
	var rowsBatch []models.URLsData
	for _, row := range shortURLsBatch {

		shortID, err := utils.HashOriginalURL([]byte(row.OriginalURL))
		if err != nil {
			logger.Log.Info(err.Error())
			return nil, fmt.Errorf("short ID creating error: %w", err)
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

	err := srv.Storage.InsertURLsDataBatch(ctx, rowsBatch)
	if err != nil {
		logger.Log.Info(err.Error())
		return nil, fmt.Errorf("inserting to db error: %w", err)
	}

	return resp, nil
}

// SendURLsToDeletion отправляет урл + id пользователя в канал для пометки урла удаленным.
func (srv *URLService) SendURLsToDeletion(urls []string, userID string) {
	for _, url := range urls {
		srv.URLsDeletionCh <- models.URLForDeleteMsg{
			ShortURL: url,
			UserID:   userID,
		}
	}
}

// flushMessages слушает сообщения в канале URLsDeletionCh, помечает, что урлы из сообщений удалены.
func (srv *URLService) flushMessages() {
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

// GetUserURLs получает неудаленные урлы по id пользователя.
func (srv *URLService) GetUserURLs(ctx context.Context, userID string) ([]models.GetUserURLsResponse, error) {
	data, err := srv.Storage.SelectURLs(ctx, userID)
	if err != nil {
		logger.Log.Info(err.Error())
		return nil, fmt.Errorf("user urls selection error: %w", err)
	}

	return data, nil
}

// GetShortURLSrv сохраняет сокращенный URL и возвращает его в качестве ответа.
func (srv *URLService) GetShortURLSrv(ctx context.Context, body []byte, userID string) (*models.ShortenURLResponse, error) {

	var jsonBody models.ShortenURLRequest

	if err := json.Unmarshal(body, &jsonBody); err != nil {
		logger.Log.Info(err.Error())
		return nil, fmt.Errorf("body unmarshal error: %w", err)
	}

	shortID, err := utils.HashOriginalURL([]byte(jsonBody.URL))
	if err != nil {
		logger.Log.Info(err.Error())
		return nil, fmt.Errorf("short ID creating error: %w", err)
	}

	resp := models.ShortenURLResponse{Result: srv.Config.BaseURL + "/" + shortID}

	event := models.URLsData{UserID: userID, UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: jsonBody.URL}

	err = srv.Storage.InsertURLsData(ctx, &event)
	if err != nil {
		logger.Log.Info(err.Error())
		return nil, err
	}

	return &resp, nil
}

// CompressURL сохраняет сокращенный URL и возвращает его в качестве ответа.
func (srv *URLService) CompressURL(ctx context.Context, originalURL []byte, userID string) (string, error) {

	shortID, err := utils.HashOriginalURL(originalURL)
	if err != nil {
		logger.Log.Info(err.Error())
		return "", fmt.Errorf("short ID creating error: %w", err)
	}

	strBody := string(originalURL)
	event := models.URLsData{UUID: uuid.New().String(), ShortURL: shortID, OriginalURL: strBody, UserID: userID}

	err = srv.Storage.InsertURLsData(ctx, &event)
	if err != nil {
		logger.Log.Info(err.Error())
		return "", err
	}

	return shortID, nil
}

// SelectOriginalURLByShortURL возвращает оригинальный урл по сокращенному.
func (srv *URLService) SelectOriginalURLByShortURL(ctx context.Context, shortURLID string) (string, error) {

	originalURL, err := srv.Storage.SelectOriginalURLByShortURL(ctx, shortURLID)
	if err != nil {
		logger.Log.Info(err.Error())
		return "", err
	}

	return originalURL, nil
}

// PingDB пингует бд.
func (srv *URLService) PingDB() error {

	err := srv.Storage.Ping()
	if err != nil {
		logger.Log.Info(err.Error())
		return err
	}

	return nil
}

// GetStats возвращает количество пользователей и урлов в сервисе.
func (srv *URLService) GetStats(ctx context.Context) (*models.GetStatsResponse, error) {

	URLsCount, err := srv.Storage.SelectURLsCount(ctx)
	if err != nil {
		logger.Log.Info("Failed to get URLs count", zap.Error(err))
		return nil, fmt.Errorf("failed to get URLs count: %w", err)
	}
	usersCount, err := srv.Storage.SelectUsersCount(ctx)
	if err != nil {
		logger.Log.Info("Failed to get users count", zap.Error(err))
		return nil, fmt.Errorf("failed to get users count: %w", err)
	}

	resp := models.GetStatsResponse{URLs: URLsCount, Users: usersCount}

	return &resp, nil
}
