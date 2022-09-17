package repositories

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

const (
	messageFileNotSpecified = "File not specified"

	fileMode = 0o777
)

type FileRepository struct {
	mu            sync.RWMutex
	shortURLs     map[models.UID]*models.ShortURL
	userShortURLs map[uuid.UUID][]*models.ShortURL
	lastID        int
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
		lastID:        0,
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
		fileRepository.lastID++
	}

	return fileRepository
}

func (f *FileRepository) ReserveID() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.lastID++

	return f.lastID
}

func (f *FileRepository) FindOneByUID(uid models.UID) *models.ShortURL {
	f.mu.RLock()
	defer f.mu.RUnlock()

	shortURL, ok := f.shortURLs[uid]
	if !ok {
		return nil
	}

	return shortURL
}

func (f *FileRepository) FindAllByUserID(uuid uuid.UUID) []*models.ShortURL {
	f.mu.RLock()
	defer f.mu.RUnlock()

	userShortURLs, ok := f.userShortURLs[uuid]
	if !ok {
		return nil
	}

	return userShortURLs
}

func (f *FileRepository) Save(shortURL *models.ShortURL) {
	f.mu.Lock()
	defer f.mu.Unlock()

	err := f.encoder.Encode(shortURL)
	if err != nil {
		log.Panic(err)
	}

	f.shortURLs[shortURL.UID] = shortURL
	f.userShortURLs[shortURL.UserID] = append(f.userShortURLs[shortURL.UserID], shortURL)
}
