package app

import (
	"fmt"
	"sync"

	"github.com/pluhe7/shortener/internal/logger"
	"go.uber.org/zap"
)

const numWorkers = 10

func (s *Server) DeleteRecords(shortIDs []string) {
	doneCh := make(chan struct{})
	defer close(doneCh)

	shortIDsCh := shortIDsChanelGenerator(doneCh, shortIDs)
	shortURLsChs := s.getShortURLsChannels(doneCh, shortIDsCh)
	shortURLsMergedCh := mergeChannels(doneCh, shortURLsChs)

	shortURLsToDelete := make([]string, 0, len(shortIDs))
	for {
		shortURL, ok := <-shortURLsMergedCh
		if ok {
			shortURLsToDelete = append(shortURLsToDelete, shortURL)

		} else {
			go func() {
				err := s.Storage.Delete(shortURLsToDelete)
				if err != nil {
					logger.Log.Error("storage delete error", zap.Error(err))
				}
			}()
			break
		}
	}
}

func (s *Server) getShortURLsChannels(doneCh chan struct{}, shortIDsCh chan string) []chan string {
	shortURLsChannels := make([]chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		shortURLsChannel := s.getShortURLToDeleteCh(doneCh, shortIDsCh)

		shortURLsChannels[i] = shortURLsChannel
	}

	return shortURLsChannels
}

func (s *Server) getShortURLToDeleteCh(doneCh chan struct{}, shortIDsCh chan string) chan string {
	shortURLsChannel := make(chan string)

	go func() {
		defer close(shortURLsChannel)

		for shortID := range shortIDsCh {
			shortURL, err := s.getURLToDelete(shortID)
			if err != nil {
				logger.Log.Error("get url to delete", zap.Error(err))
				continue
			}

			select {
			case <-doneCh:
				return
			case shortURLsChannel <- shortURL:
			}
		}
	}()

	return shortURLsChannel
}

func (s *Server) getURLToDelete(shortID string) (string, error) {
	shortURL := s.getShortURLFromID(shortID)

	record, err := s.Storage.Get(shortURL)
	if err != nil {
		return "", fmt.Errorf("get record: %w", err)
	}

	if record.UserID != s.SessionUserID {
		return "", fmt.Errorf("unable delete record with id %s for user %s", shortID, s.SessionUserID)
	}

	return shortURL, nil
}

func shortIDsChanelGenerator(doneCh chan struct{}, shortIDs []string) chan string {
	shortIDsCh := make(chan string)

	go func() {
		defer close(shortIDsCh)

		for _, id := range shortIDs {
			select {
			case <-doneCh:
				return
			case shortIDsCh <- id:
			}
		}
	}()

	return shortIDsCh
}

func mergeChannels(doneCh chan struct{}, resultChs []chan string) chan string {
	finalCh := make(chan string)

	var wg sync.WaitGroup
	wg.Add(len(resultChs))

	for _, ch := range resultChs {
		chClosure := ch

		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
