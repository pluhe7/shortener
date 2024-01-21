package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/pluhe7/shortener/internal/logger"
	"github.com/pluhe7/shortener/internal/models"
)

type ShortURLStorage struct {
	ShortURLs map[string]string
	filename  string
	Database  *sql.DB
}

func NewShortURLStorage(filename, databaseDSN string) (*ShortURLStorage, error) {
	storage := ShortURLStorage{
		ShortURLs: make(map[string]string),
		filename:  filename,
	}

	if databaseDSN != "" {
		db, err := sql.Open("pgx", databaseDSN)
		if err != nil {
			return nil, fmt.Errorf("open db connection: %w", err)
		}
		storage.Database = db
	}

	if filename != "" {
		if err := storage.restoreFromFile(); err != nil {
			return nil, fmt.Errorf("restore storage from file: %w", err)
		}
	}

	return &storage, nil
}

func (s *ShortURLStorage) Add(shortURL, fullURL string) error {
	record := models.ShortURLRecord{
		ID:          len(s.ShortURLs) + 1,
		ShortURL:    shortURL,
		OriginalURL: fullURL,
	}

	if s.filename != "" {
		w, err := newDataWriter(s.filename)
		if err != nil {
			return fmt.Errorf("new data writer: %w", err)
		}
		defer w.Close()

		err = w.WriteData(&record)
		if err != nil {
			return fmt.Errorf("write data: %w", err)
		}
	}

	s.ShortURLs[shortURL] = fullURL

	return nil
}

func (s *ShortURLStorage) restoreFromFile() error {
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

func (s *ShortURLStorage) Get(shortURL string) (string, bool) {
	fullURL, ok := s.ShortURLs[shortURL]

	return fullURL, ok
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
