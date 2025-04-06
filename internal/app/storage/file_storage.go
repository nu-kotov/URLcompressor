package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type FileStorage struct {
	dataProducer *Producer
	dataConsumer *Consumer
	mapCash      map[string]string
}

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

func (f *FileStorage) InsertURLsData(ctx context.Context, data *models.URLsData) error {
	f.mapCash[data.ShortURL] = data.OriginalURL
	return f.dataProducer.WriteEvent(data)
}

func (f *FileStorage) DeleteURLs(ctx context.Context, data []models.URLForDeleteMsg) error {
	return nil
}

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

func (f *FileStorage) SelectOriginalURLByShortURL(ctx context.Context, shortURL string) (string, error) {
	if _, exist := f.mapCash[shortURL]; !exist {
		return "", errors.New("SHORT URL NOT EXIST")
	}
	return f.mapCash[shortURL], nil
}

func (f *FileStorage) Ping() error {
	return nil
}

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

func (c *Consumer) Close() error {
	return c.file.Close()
}
