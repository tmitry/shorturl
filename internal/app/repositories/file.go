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

func (f *FileRepository) Save(
	_ context.Context,
	url models.URL,
	userID uuid.UUID,
	hashMinLength int,
	hashSalt string,
) (*models.ShortURL, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	id := len(f.shortURLs) + 1

	shortURL := models.NewShortURL(id, url, models.NewUID(id, hashMinLength, hashSalt), userID)

	err := f.encoder.Encode(shortURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	f.shortURLs[shortURL.UID] = shortURL
	f.userShortURLs[shortURL.UserID] = append(f.userShortURLs[shortURL.UserID], shortURL)

	return shortURL, nil
}

func (f *FileRepository) Ping(_ context.Context) error {
	return nil
}
