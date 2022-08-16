package repositories

import (
	"github.com/tmitry/shorturl/internal/app/models"
)

type MemoryRepository struct {
	shortURLs map[models.UID]models.ShortURL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		shortURLs: map[models.UID]models.ShortURL{},
	}
}

func (m MemoryRepository) Find(u models.UID) *models.ShortURL {
	shortURL, ok := m.shortURLs[u]
	if !ok {
		return nil
	}

	return &shortURL
}

func (m *MemoryRepository) Add(u models.URL) *models.ShortURL {
	shortURL := models.NewShortURL(len(m.shortURLs), u)

	m.shortURLs[shortURL.UID] = *shortURL

	return shortURL
}
