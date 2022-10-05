package repositories

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

type MemoryRepository struct {
	mu            sync.RWMutex
	shortURLs     map[models.UID]*models.ShortURL
	userShortURLs map[uuid.UUID][]*models.ShortURL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		mu:            sync.RWMutex{},
		shortURLs:     map[models.UID]*models.ShortURL{},
		userShortURLs: map[uuid.UUID][]*models.ShortURL{},
	}
}

func (m *MemoryRepository) FindOneByUID(_ context.Context, uid models.UID) (*models.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	shortURL, ok := m.shortURLs[uid]
	if !ok {
		return nil, ErrNotFound
	}

	return shortURL, nil
}

func (m *MemoryRepository) FindAllByUserID(_ context.Context, userID uuid.UUID) ([]*models.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	userShortURLs, ok := m.userShortURLs[userID]
	if !ok {
		return nil, ErrNotFound
	}

	return userShortURLs, nil
}

func (m *MemoryRepository) Save(
	_ context.Context,
	url models.URL,
	userID uuid.UUID,
	hashMinLength int,
	hashSalt string,
) (*models.ShortURL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := len(m.shortURLs) + 1

	shortURL := models.NewShortURL(id, url, models.NewUID(id, hashMinLength, hashSalt), userID)

	m.shortURLs[shortURL.UID] = shortURL
	m.userShortURLs[shortURL.UserID] = append(m.userShortURLs[shortURL.UserID], shortURL)

	return shortURL, nil
}

func (m *MemoryRepository) BatchSave(
	_ context.Context,
	urls []models.URL,
	userID uuid.UUID,
	hashMinLength int,
	hashSalt string,
) ([]*models.ShortURL, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	batchShortURLs := make([]*models.ShortURL, 0, len(urls))

	for _, url := range urls {
		id := len(m.shortURLs) + 1
		shortURL := models.NewShortURL(id, url, models.NewUID(id, hashMinLength, hashSalt), userID)
		batchShortURLs = append(batchShortURLs, shortURL)

		m.shortURLs[shortURL.UID] = shortURL
		m.userShortURLs[shortURL.UserID] = append(m.userShortURLs[shortURL.UserID], shortURL)
	}

	return batchShortURLs, nil
}

func (m *MemoryRepository) Ping(_ context.Context) error {
	return nil
}
