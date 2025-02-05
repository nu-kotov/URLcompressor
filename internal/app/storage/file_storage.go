package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/nu-kotov/URLcompressor/internal/app/models"
)

type FileStorage struct {
	DataProducer *Producer
	DataConsumer *Consumer
}

func NewFileStorage(filename string) (*FileStorage, error) {
	producer, err := newProducer(filename)
	if err != nil {
		return nil, err
	}
	consumer, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}
	return &FileStorage{
		DataProducer: producer,
		DataConsumer: consumer,
	}, nil
}

func (f *FileStorage) ProduceEvent(event *models.URLsData) error {
	return f.DataProducer.WriteEvent(event)
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

func (c *Consumer) Close() error {
	return c.file.Close()
}

func InitMapStorage(c *Consumer) (map[string]string, error) {

	mapStorage := make(map[string]string)
	for {
		fileStr, err := c.ReadEvent()
		if err != nil {
			return nil, err
		}
		if fileStr == nil {
			break
		}
		mapStorage[fileStr.ShortURL] = fileStr.OriginalURL

	}
	c.Close()

	return mapStorage, nil
}
