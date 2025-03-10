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
	Ping() error
	Close() error
}

func NewStorage(c config.Config) (Storage, error) {
	if c.FileStoragePath != "" {
		fileStorage, err := NewFileStorage(c.FileStoragePath)
		if err != nil {
			return nil, err
		}

		return fileStorage, nil
	} else if c.DatabaseConnection != "" {
		DBStorage, err := NewConnect(c.DatabaseConnection)
		if err != nil {
			return nil, err
		}

		err = DBStorage.CreateTable()
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
