package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

func (s *FileStorage) Get(shortURL string) (*models.ShortURLRecord, error) {
	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		record, err := s.readRecord(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read record: %w", err)
		}

		if record.ShortURL == shortURL {
			return record, nil
		}
	}

	return nil, ErrURLNotFound
}

func (s *FileStorage) GetByOriginal(originalURL string) (*models.ShortURLRecord, error) {
	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		record, err := s.readRecord(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read record: %w", err)
		}

		if record.OriginalURL == originalURL {
			return record, nil
		}
	}

	return nil, ErrURLNotFound
}

func (s *FileStorage) FindByUserID(userID string) ([]models.ShortURLRecord, error) {
	var records []models.ShortURLRecord

	file, err := os.OpenFile(s.filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		record, err := s.readRecord(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read record: %w", err)
		}

		if record.UserID == userID {
			records = append(records, *record)
		}
	}

	return records, nil
}

func (s *FileStorage) readRecord(reader *bufio.Reader) (*models.ShortURLRecord, error) {
	recordBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read record bytes: %w", err)
	}

	record := &models.ShortURLRecord{}
	err = json.Unmarshal(recordBytes, record)
	if err != nil {
		return nil, fmt.Errorf("unmarshal record: %w", err)
	}

	return record, nil
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

func (s *FileStorage) Delete(shortURLs []string) error {
	return nil
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
