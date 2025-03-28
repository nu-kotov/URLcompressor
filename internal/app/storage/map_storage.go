package storage

import (
	"context"
	"errors"

	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type MapStorage struct {
	mapStorage map[string]string
}

func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{
		mapStorage: make(map[string]string),
	}, nil
}

func (ms *MapStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	if _, exist := ms.mapStorage[data.ShortURL]; !exist {
		ms.mapStorage[data.ShortURL] = data.OriginalURL
	}
	return nil
}

func (ms *MapStorage) InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error {
	for _, d := range data {
		if _, exist := ms.mapStorage[d.ShortURL]; !exist {
			ms.mapStorage[d.ShortURL] = d.OriginalURL
		}
	}
	return nil
}

func (ms *MapStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if _, exist := ms.mapStorage[shortURL]; !exist {
		return "", errors.New("SHORT URL NOT EXIST")
	}
	return ms.mapStorage[shortURL], nil
}

func (ms *MapStorage) SelectURLs(ctx context.Context, userID string) ([]models.GetUserURLsResponse, error) {
	var data []models.GetUserURLsResponse

	for su, ou := range ms.mapStorage {
		data = append(data, models.GetUserURLsResponse{ShortURL: su, OriginalURL: ou})
	}

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

func (ms *MapStorage) Ping() error {
	return nil
}

func (ms *MapStorage) Close() error {
	return nil
}
