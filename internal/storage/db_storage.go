package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/pluhe7/shortener/internal/models"
)

var ErrDuplicateRecord = errors.New("url already exist")

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

func (s *DatabaseStorage) Get(shortURL string) (*models.ShortURLRecord, error) {
	row := s.db.QueryRow("SELECT short_url, original_url, user_id FROM urls WHERE short_url = $1", shortURL)

	var record models.ShortURLRecord
	err := row.Scan(&record.ShortURL, &record.OriginalURL, &record.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("scan original url: %w", err)
	}

	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("row err: %w", err)
	}

	return &record, nil
}

func (s *DatabaseStorage) GetByOriginal(originalURL string) (*models.ShortURLRecord, error) {
	row := s.db.QueryRow("SELECT short_url, original_url, user_id FROM urls WHERE original_url = $1", originalURL)

	var record models.ShortURLRecord
	err := row.Scan(&record.ShortURL, &record.OriginalURL, &record.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrURLNotFound
		}
		return nil, fmt.Errorf("scan original url: %w", err)
	}

	if err = row.Err(); err != nil {
		return nil, fmt.Errorf("row err: %w", err)
	}

	return &record, nil
}

func (s *DatabaseStorage) FindByUserID(userID string) ([]models.ShortURLRecord, error) {
	rows, err := s.db.Query("SELECT short_url, original_url, user_id FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var records []models.ShortURLRecord
	for rows.Next() {
		var record models.ShortURLRecord

		err = rows.Scan(&record.ShortURL, &record.OriginalURL, &record.UserID)
		if err != nil {
			return nil, fmt.Errorf("scan record: %w", err)
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return records, nil
}

func (s *DatabaseStorage) Save(record models.ShortURLRecord) error {
	res, err := s.db.Exec(`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		record.ShortURL, record.OriginalURL, record.UserID)
	if err != nil {
		return fmt.Errorf("insert: %w", err)
	}

	insertedRowsCount, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get inserted rows count: %w", err)
	}

	if insertedRowsCount != 1 {
		return ErrDuplicateRecord
	}

	return nil
}

func (s *DatabaseStorage) SaveBatch(records []models.ShortURLRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)  ON CONFLICT DO NOTHING")
	if err != nil {
		return fmt.Errorf("prepare sql: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		_, err = stmt.Exec(record.ShortURL, record.OriginalURL, record.UserID)
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
		 short_url VARCHAR(255) PRIMARY KEY,
		 original_url TEXT NOT NULL UNIQUE,
		 user_id VARCHAR(255)
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
