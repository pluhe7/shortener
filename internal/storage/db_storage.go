package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/pluhe7/shortener/internal/models"
)

type DatabaseStorage struct {
	db *sql.DB
}

func NewDatabaseStorage(dsn string) (*DatabaseStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db connection: %w", err)
	}

	s := &DatabaseStorage{
		db: db,
	}

	err = s.migrateURLsTable()
	if err != nil {
		return nil, fmt.Errorf("migrate urls table: %w", err)
	}

	return s, nil
}

func (s *DatabaseStorage) Get(shortURL string) (string, error) {
	row := s.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL)

	var originalURL string
	err := row.Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		} else {
			return "", fmt.Errorf("scan original url: %w", err)
		}
	}

	return originalURL, nil
}

func (s *DatabaseStorage) Save(record models.ShortURLRecord) error {
	_, err := s.db.Exec(`INSERT INTO urls (short_url, original_url) VALUES ($1, $2)`,
		record.ShortURL, record.OriginalURL)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	return nil
}

func (s *DatabaseStorage) SaveBatch(records []models.ShortURLRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)")
	if err != nil {
		return fmt.Errorf("prepare sql: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err = stmt.Exec(record.ShortURL, record.OriginalURL)
		if err != nil {
			return fmt.Errorf("insert short %s for original %s error: %w", record.ShortURL, record.OriginalURL, err)
		}
	}

	return tx.Commit()
}

func (s *DatabaseStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}

	return nil
}

func (s *DatabaseStorage) migrateURLsTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		 id SERIAL PRIMARY KEY,
		 short_url VARCHAR(255) NOT NULL UNIQUE,
		 original_url TEXT NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("execute create table query: %w", err)
	}

	return nil
}

func (s *DatabaseStorage) PingContext(ctx context.Context) error {
	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
