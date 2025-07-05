package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

// FileStorage - структура хранилища в файле.
type FileStorage struct {
	dataProducer *Producer
	dataConsumer *Consumer
	mapCash      map[string]string
}

// NewFileStorage - конструктор хранилища в файле.
func NewFileStorage(filename string, baseURL string) (*FileStorage, error) {
	producer, err := newProducer(filename)
	if err != nil {
		return nil, err
	}

	consumer, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}

	cash, err := consumer.fillMapCash()
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		dataProducer: producer,
		dataConsumer: consumer,
		mapCash:      cash,
	}, nil
}

// InsertURLsData - вставляет в файл информацию по урлу.
func (f *FileStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	f.mapCash[data.ShortURL] = data.OriginalURL
	return f.dataProducer.WriteEvent(data)
}

// DeleteURLs - заглушка, для реализации общего интерфейса для всех видов хранилищ.
func (f *FileStorage) DeleteURLs(ctx context.Context, data []models.URLForDeleteMsg) error {
	return nil
}

// InsertURLsDataBatch - вставка батча урлов в файл.
func (f *FileStorage) InsertURLsDataBatch(ctx context.Context, data []models.URLsData) error {
	for _, d := range data {
		if _, exist := f.mapCash[d.ShortURL]; !exist {
			f.mapCash[d.ShortURL] = d.OriginalURL
			err := f.dataProducer.WriteEvent(&d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SelectURLs - возвращает информацию по урлам пользователя из файла.
func (f *FileStorage) SelectURLs(ctx context.Context, userID string) ([]models.GetUserURLsResponse, error) {
	var data []models.GetUserURLsResponse

	for {
		fileStr, err := f.dataConsumer.ReadEvent()
		if err != nil {
			return nil, err
		}
		if fileStr == nil {
			break
		}
		data = append(data, models.GetUserURLsResponse{ShortURL: fileStr.ShortURL, OriginalURL: fileStr.OriginalURL})
	}

	if len(data) == 0 {
		return nil, ErrNotFound
	}

	return data, nil
}

// SelectOriginalURLByShortURL - возвращает полный урл по сокращенному из файла.
func (f *FileStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if _, exist := f.mapCash[shortURL]; !exist {
		return "", errors.New("SHORT URL NOT EXIST")
	}
	return f.mapCash[shortURL], nil
}

// Ping - заглушка, для реализации общего интерфейса для всех видов хранилищ.
func (f *FileStorage) Ping() error {
	return nil
}

// Close - вызывает методы закрытия файла консюмера и продюсера.
func (f *FileStorage) Close() error {
	err := f.dataConsumer.file.Close()
	if err != nil {
		return err
	}

	err = f.dataProducer.file.Close()
	if err != nil {
		return err
	}
	return nil
}

// Producer - экземпляр продюсера для записи в файл.
type Producer struct {
	file   *os.File
	writer *bufio.Writer
}

func newProducer(filename string) (*Producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Producer{
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// WriteEvent - записывает данные в файл.
func (p *Producer) WriteEvent(event *models.URLsData) error {
	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	return p.writer.Flush()
}

// Consumer - экземпляр консюмера для чтения из файла.
type Consumer struct {
	file    *os.File
	scanner *bufio.Scanner
}

func newConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		file:    file,
		scanner: bufio.NewScanner(file),
	}, nil
}

// ReadEvent - читает данные из файла.
func (c *Consumer) ReadEvent() (*models.URLsData, error) {

	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	data := c.scanner.Bytes()

	event := models.URLsData{}
	err := json.Unmarshal(data, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (c *Consumer) fillMapCash() (map[string]string, error) {

	mapCash := make(map[string]string)
	for {
		fileStr, err := c.ReadEvent()
		if err != nil {
			return nil, err
		}
		if fileStr == nil {
			break
		}
		mapCash[fileStr.ShortURL] = fileStr.OriginalURL

	}

	return mapCash, nil
}

// Close - закрывает файл.
func (c *Consumer) Close() error {
	return c.file.Close()
}
