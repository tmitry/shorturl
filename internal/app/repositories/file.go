package repositories

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/tmitry/shorturl/internal/app/config"
	"github.com/tmitry/shorturl/internal/app/models"
)

const (
	messageFileNotSpecified = "File not specified"

	fileMode = 0o777
)

type FileRepository struct {
	mu        sync.Mutex
	shortURLs map[models.UID]*models.ShortURL
	lastID    int
	encoder   *json.Encoder
}

func NewFileRepository() *FileRepository {
	if config.AppCfg.FileStoragePath == "" {
		log.Panic(messageFileNotSpecified)
	}

	fileWriter, err := os.OpenFile(config.AppCfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, fileMode)
	if err != nil {
		log.Panic(err)
	}

	encoder := json.NewEncoder(fileWriter)
	encoder.SetEscapeHTML(false)

	fileRepository := &FileRepository{
		mu:        sync.Mutex{},
		shortURLs: map[models.UID]*models.ShortURL{},
		lastID:    0,
		encoder:   encoder,
	}

	fileReader, err := os.OpenFile(config.AppCfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, fileMode)
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
		shortURL := models.NewShortURL(0, "", "")
		if err := decoder.Decode(shortURL); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			log.Panic(err)
		}

		fileRepository.shortURLs[shortURL.UID] = shortURL
		fileRepository.lastID++
	}

	return fileRepository
}

func (f *FileRepository) ReserveID() int {
	f.mu.Lock()
	f.lastID++
	defer f.mu.Unlock()

	return f.lastID
}

func (f *FileRepository) Find(uid models.UID) *models.ShortURL {
	shortURL, ok := f.shortURLs[uid]
	if !ok {
		return nil
	}

	return shortURL
}

func (f *FileRepository) Save(shortURL *models.ShortURL) {
	err := f.encoder.Encode(shortURL)
	if err != nil {
		log.Panic(err)
	}

	f.shortURLs[shortURL.UID] = shortURL
}
