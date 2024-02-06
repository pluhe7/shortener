package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pluhe7/shortener/internal/models"
)

type FileStorage struct {
	filename string
}

func NewFileStorage(filename string) (*FileStorage, error) {
	storage := FileStorage{
		filename: filename,
	}

	return &storage, nil
}

func (s *FileStorage) Get(shortURL string) (string, error) {
	var originalURL string

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		recordBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read record bytes: %w", err)
		}

		var record models.ShortURLRecord
		err = json.Unmarshal(recordBytes, &record)
		if err != nil {
			return "", fmt.Errorf("unmarshal record: %w", err)
		}

		if record.ShortURL == shortURL {
			originalURL = record.OriginalURL
			break
		}
	}

	if originalURL == "" {
		return "", ErrURLNotFound
	}

	return originalURL, nil
}

func (s *FileStorage) GetByOriginal(originalURL string) (string, error) {
	var shortURL string

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		recordBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("read record bytes: %w", err)
		}

		var record models.ShortURLRecord
		err = json.Unmarshal(recordBytes, &record)
		if err != nil {
			return "", fmt.Errorf("unmarshal record: %w", err)
		}

		if record.OriginalURL == originalURL {
			shortURL = record.ShortURL
			break
		}
	}

	if shortURL == "" {
		return "", ErrURLNotFound
	}

	return shortURL, nil
}

func (s *FileStorage) Save(record models.ShortURLRecord) error {
	recordCount, err := s.getRecordCount()
	if err != nil {
		return fmt.Errorf("get record count: %w", err)
	}

	record.ID = recordCount + 1

	w, err := newDataWriter(s.filename)
	if err != nil {
		return fmt.Errorf("new data writer: %w", err)
	}
	defer w.Close()

	err = w.WriteData(&record)
	if err != nil {
		return fmt.Errorf("write data: %w", err)
	}

	return nil
}

func (s *FileStorage) SaveBatch(records []models.ShortURLRecord) error {
	w, err := newDataWriter(s.filename)
	if err != nil {
		return fmt.Errorf("new data writer: %w", err)
	}
	defer w.Close()

	recordCount, err := s.getRecordCount()
	if err != nil {
		return fmt.Errorf("get record count: %w", err)
	}

	for i, record := range records {
		record.ID = recordCount + i + 1

		err = w.WriteData(&record)
		if err != nil {
			return fmt.Errorf("write data: %w", err)
		}
	}

	return nil
}

func (s *FileStorage) getRecordCount() (int, error) {
	var recordCount int

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return 0, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		_, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, fmt.Errorf("read bytes: %w", err)
		}

		recordCount++
	}

	return recordCount, nil
}

func (s *FileStorage) Close() error {
	return nil
}

func (s *FileStorage) PingContext(ctx context.Context) error {
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
