package storage

import (
	"context"

	"github.com/nu-kotov/URLcompressor/config"
	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type Storage interface {
	InsertURLsData(ctx context.Context, data *models.URLsData) error
	InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error
	SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error)
	SelectURLs(ctx context.Context, userID string) ([]models.GetUserURLsResponse, error)
	DeleteURLs(ctx context.Context, data []models.URLForDeleteMsg) error
	Ping() error
	Close() error
}

func NewStorage(c config.Config) (Storage, error) {
	if c.FileStoragePath != "" {
		fileStorage, err := NewFileStorage(c.FileStoragePath, c.BaseURL)
		if err != nil {
			return nil, err
		}

		return fileStorage, nil

	} else if c.DatabaseConnection != "" {
		DBStorage, err := NewConnect(c.DatabaseConnection, c.BaseURL)
		if err != nil {
			return nil, err
		}

		return DBStorage, nil

	} else {
		mapStorage, err := NewMapStorage()
		if err != nil {
			return nil, err
		}

		return mapStorage, nil

	}
}
