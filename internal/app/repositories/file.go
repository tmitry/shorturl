package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

const (
	messageFileNotSpecified = "file not specified"

	fileMode = 0o777
)

type FileRepository struct {
	mu            sync.RWMutex
	shortURLs     map[models.UID]*models.ShortURL
	userShortURLs map[uuid.UUID][]*models.ShortURL
	encoder       *json.Encoder
}

func NewFileRepository(fileStoragePath string) *FileRepository {
	if fileStoragePath == "" {
		log.Panic(messageFileNotSpecified)
	}

	fileWriter, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		log.Panic(err)
	}

	encoder := json.NewEncoder(fileWriter)
	encoder.SetEscapeHTML(false)

	fileRepository := &FileRepository{
		mu:            sync.RWMutex{},
		shortURLs:     map[models.UID]*models.ShortURL{},
		userShortURLs: map[uuid.UUID][]*models.ShortURL{},
		encoder:       encoder,
	}

	fileReader, err := os.OpenFile(fileStoragePath, os.O_RDONLY|os.O_CREATE, fileMode)
	if err != nil {
		log.Panic(err)
	}

	defer func(fileReader *os.File) {
		err := fileReader.Close()
		if err != nil {
			log.Panic(err)
		}
	}(fileReader)

	decoder := json.NewDecoder(fileReader)

	for {
		shortURL := models.NewShortURL(0, "", "", uuid.UUID{})
		if err := decoder.Decode(shortURL); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Panic(err)
		}

		fileRepository.shortURLs[shortURL.UID] = shortURL
		fileRepository.userShortURLs[shortURL.UserID] = append(fileRepository.userShortURLs[shortURL.UserID], shortURL)
	}

	return fileRepository
}

func (f *FileRepository) FindOneByUID(_ context.Context, uid models.UID) (*models.ShortURL, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	shortURL, ok := f.shortURLs[uid]
	if !ok {
		return nil, ErrNotFound
	}

	return shortURL, nil
}

func (f *FileRepository) FindAllByUserID(_ context.Context, userID uuid.UUID) ([]*models.ShortURL, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	userShortURLs, ok := f.userShortURLs[userID]
	if !ok {
		return nil, ErrNotFound
	}

	return userShortURLs, nil
}

func (f *FileRepository) Save(_ context.Context, shortURL *models.ShortURL) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	userShortURLs, ok := f.userShortURLs[shortURL.UserID]
	if ok {
		for _, userShortURL := range userShortURLs {
			if userShortURL.URL == shortURL.URL {
				*shortURL = *userShortURL

				return ErrURLDuplicate
			}
		}
	}

	err := f.encoder.Encode(shortURL)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	f.shortURLs[shortURL.UID] = shortURL
	f.userShortURLs[shortURL.UserID] = append(f.userShortURLs[shortURL.UserID], shortURL)

	return nil
}

func (f *FileRepository) BatchSave(_ context.Context, shortURLs []*models.ShortURL) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, shortURL := range shortURLs {
		isFound := false

		userShortURLs, ok := f.userShortURLs[shortURL.UserID]
		if ok {
			for _, userShortURL := range userShortURLs {
				if userShortURL.URL == shortURL.URL {
					*shortURL = *userShortURL
					isFound = true

					break
				}
			}
		}

		if isFound {
			continue
		}

		err := f.encoder.Encode(shortURL)
		if err != nil {
			log.Panic(err)
		}

		f.shortURLs[shortURL.UID] = shortURL
		f.userShortURLs[shortURL.UserID] = append(f.userShortURLs[shortURL.UserID], shortURL)
	}

	return nil
}

func (f *FileRepository) BatchDelete(_ context.Context, shortURLs []*models.ShortURL) error {
	if len(shortURLs) == 0 {
		return ErrNothingToDelete
	}

	// not implemented
	for _, shortURL := range shortURLs {
		shortURL.IsDeleted = true
	}

	return nil
}

func (f *FileRepository) FindAllByUserIDAndUIDs(
	_ context.Context,
	userID uuid.UUID,
	uids []models.UID,
) ([]*models.ShortURL, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var shortURLs []*models.ShortURL

	for _, uid := range uids {
		shortURL, ok := f.shortURLs[uid]
		if ok && shortURL.UserID == userID {
			shortURLs = append(shortURLs, shortURL)
		}
	}

	if len(shortURLs) == 0 {
		return nil, ErrNotFound
	}

	return shortURLs, nil
}

func (f *FileRepository) Ping(_ context.Context) error {
	return nil
}
