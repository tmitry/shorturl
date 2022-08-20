package repositories

import (
	"github.com/tmitry/shorturl/internal/app/models"
)

type MemoryRepository struct {
	shortURLs map[models.UID]models.ShortURL
	lastID    int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		shortURLs: map[models.UID]models.ShortURL{},
		lastID:    0,
	}
}

func (m *MemoryRepository) ReserveID() int {
	m.lastID++

	return m.lastID
}

func (m MemoryRepository) Find(uid models.UID) *models.ShortURL {
	shortURL, ok := m.shortURLs[uid]
	if !ok {
		return nil
	}

	return &shortURL
}

func (m *MemoryRepository) Add(shortURL *models.ShortURL) {
	m.shortURLs[shortURL.UID] = *shortURL
}
