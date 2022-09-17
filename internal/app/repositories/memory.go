package repositories

import (
	"sync"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

type MemoryRepository struct {
	mu            sync.Mutex
	shortURLs     map[models.UID]*models.ShortURL
	userShortURLs map[uuid.UUID][]*models.ShortURL
	lastID        int
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mu:            sync.Mutex{},
		shortURLs:     map[models.UID]*models.ShortURL{},
		userShortURLs: map[uuid.UUID][]*models.ShortURL{},
		lastID:        0,
	}
}

func (m *MemoryRepository) ReserveID() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastID++

	return m.lastID
}

func (m *MemoryRepository) FindOneByUID(uid models.UID) *models.ShortURL {
	shortURL, ok := m.shortURLs[uid]
	if !ok {
		return nil
	}

	return shortURL
}

func (m *MemoryRepository) FindAllByUserID(uuid uuid.UUID) []*models.ShortURL {
	userShortURLs, ok := m.userShortURLs[uuid]
	if !ok {
		return nil
	}

	return userShortURLs
}

func (m *MemoryRepository) Save(shortURL *models.ShortURL) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.shortURLs[shortURL.UID] = shortURL
	m.userShortURLs[shortURL.UserID] = append(m.userShortURLs[shortURL.UserID], shortURL)
}
