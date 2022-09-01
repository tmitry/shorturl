package repositories

import (
	"sync"

	"github.com/tmitry/shorturl/internal/app/models"
)

type MemoryRepository struct {
	mu        sync.Mutex
	shortURLs map[models.UID]*models.ShortURL
	lastID    int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mu:        sync.Mutex{},
		shortURLs: map[models.UID]*models.ShortURL{},
		lastID:    0,
	}
}

func (m *MemoryRepository) ReserveID() int {
	m.mu.Lock()
	m.lastID++
	defer m.mu.Unlock()

	return m.lastID
}

func (m *MemoryRepository) Find(uid models.UID) *models.ShortURL {
	shortURL, ok := m.shortURLs[uid]
	if !ok {
		return nil
	}

	return shortURL
}

func (m *MemoryRepository) Save(shortURL *models.ShortURL) {
	m.shortURLs[shortURL.UID] = shortURL
}
