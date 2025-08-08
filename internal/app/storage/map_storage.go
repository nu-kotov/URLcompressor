package storage

import (
	"context"
	"errors"

	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

// MapStorage - структура хранилища в памяти.
type MapStorage struct {
	mapStorage map[string]string
}

// NewMapStorage - конструктор хранилища в памяти.
func NewMapStorage() (*MapStorage, error) {
	return &MapStorage{
		mapStorage: make(map[string]string),
	}, nil
}

// SelectURLsCount - получает количество урлов в сервисе.
func (ms *MapStorage) SelectURLsCount(ctx context.Context) (int, error) {
	return len(ms.mapStorage), nil
}

// SelectURLsCount - заглушка метода получения количество пользователей в сервисе.
func (ms *MapStorage) SelectUsersCount(ctx context.Context) (int, error) {
	return -2, nil
}

// InsertURLsData - вставляет в мапу информацию по урлу.
func (ms *MapStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	if _, exist := ms.mapStorage[data.ShortURL]; !exist {
		ms.mapStorage[data.ShortURL] = data.OriginalURL
	}
	return nil
}

// InsertURLsDataBatch - вставляет в мапу батч урлов.
func (ms *MapStorage) InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error {
	for _, d := range data {
		if _, exist := ms.mapStorage[d.ShortURL]; !exist {
			ms.mapStorage[d.ShortURL] = d.OriginalURL
		}
	}
	return nil
}

// SelectOriginalURLByShortURL - возвращает полный урл по сокращенному из мапы.
func (ms *MapStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if _, exist := ms.mapStorage[shortURL]; !exist {
		return "", errors.New("SHORT URL NOT EXIST")
	}
	return ms.mapStorage[shortURL], nil
}

// SelectURLs - возвращает полный урл по сокращенному из мапы.
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

// DeleteURLs - заглушка, для реализации общего интерфейса для всех видов хранилищ.
func (ms *MapStorage) DeleteURLs(ctx context.Context, data []models.URLForDeleteMsg) error {
	return nil
}

// Ping - заглушка, для реализации общего интерфейса для всех видов хранилищ.
func (ms *MapStorage) Ping() error {
	return nil
}

// Close - заглушка, для реализации общего интерфейса для всех видов хранилищ.
func (ms *MapStorage) Close() error {
	return nil
}
