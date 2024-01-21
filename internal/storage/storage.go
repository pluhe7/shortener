package storage

import (
	"context"
	"fmt"
)

var (
	ErrURLNotFound = fmt.Errorf("url does not exist")
)

type Storage interface {
	Get(shortURL string) (string, error)
	Add(shortURL string, originalURL string) error
	Close() error
	PingContext(ctx context.Context) error
}

func NewStorage(storageFilename, databaseDSN string) (Storage, error) {
	var s Storage
	var err error

	if databaseDSN != "" {
		s, err = NewDatabaseStorage(databaseDSN)
		if err != nil {
			return nil, fmt.Errorf("new db storage: %w", err)
		}

	} else if storageFilename != "" {
		s, err = NewFileStorage(storageFilename)
		if err != nil {
			return nil, fmt.Errorf("new file storage: %w", err)
		}

	} else {
		s, err = NewMemoryStorage()
		if err != nil {
			return nil, fmt.Errorf("new memory storage: %w", err)
		}
	}

	return s, nil
}
