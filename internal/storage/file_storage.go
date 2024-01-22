package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/pluhe7/shortener/internal/logger"
	"github.com/pluhe7/shortener/internal/models"
)

type FileStorage struct {
	ShortURLs map[string]string
	filename  string
}

func NewFileStorage(filename string) (*FileStorage, error) {
	storage := FileStorage{
		ShortURLs: make(map[string]string),
		filename:  filename,
	}

	if filename != "" {
		if err := storage.restoreFromFile(); err != nil {
			return nil, fmt.Errorf("restore storage from file: %w", err)
		}
	}

	return &storage, nil
}

func (s *FileStorage) Get(shortURL string) (string, error) {
	originalURL, ok := s.ShortURLs[shortURL]
	if !ok {
		return "", ErrURLNotFound
	}

	return originalURL, nil
}

func (s *FileStorage) Save(record models.ShortURLRecord) error {
	record.ID = len(s.ShortURLs) + 1

	w, err := newDataWriter(s.filename)
	if err != nil {
		return fmt.Errorf("new data writer: %w", err)
	}
	defer w.Close()

	err = w.WriteData(&record)
	if err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	s.ShortURLs[record.ShortURL] = record.OriginalURL

	return nil
}

func (s *FileStorage) SaveBatch(records []models.ShortURLRecord) error {
	w, err := newDataWriter(s.filename)
	if err != nil {
		return fmt.Errorf("new data writer: %w", err)
	}
	defer w.Close()

	for _, record := range records {
		record.ID = len(s.ShortURLs) + 1

		s.ShortURLs[record.ShortURL] = record.OriginalURL

		err = w.WriteData(&record)
		if err != nil {
			return fmt.Errorf("write data: %w", err)
		}

		s.ShortURLs[record.ShortURL] = record.OriginalURL
	}

	return nil
}

func (s *FileStorage) Close() error {
	return nil
}

func (s *FileStorage) PingContext(ctx context.Context) error {
	return nil
}

func (s *FileStorage) restoreFromFile() error {
	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	for fileScanner.Scan() {
		var record models.ShortURLRecord

		row := fileScanner.Text()
		err = json.Unmarshal([]byte(row), &record)
		if err != nil {
			logger.Log.Error("unmarshal row error", zap.Error(err))
			continue
		}

		s.ShortURLs[record.ShortURL] = record.OriginalURL
	}

	return nil
}

type dataWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func newDataWriter(fileName string) (*dataWriter, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return &dataWriter{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (w *dataWriter) WriteData(record *models.ShortURLRecord) error {
	return w.encoder.Encode(record)
}

func (w *dataWriter) Close() error {
	return w.file.Close()
}
