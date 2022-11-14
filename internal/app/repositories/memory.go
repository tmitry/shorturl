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

func (m *MemoryRepository) Save(_ context.Context, shortURL *models.ShortURL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	userShortURLs, ok := m.userShortURLs[shortURL.UserID]
	if ok {
		for _, userShortURL := range userShortURLs {
			if userShortURL.URL == shortURL.URL {
				*shortURL = *userShortURL

				return ErrURLDuplicate
			}
		}
	}

	m.shortURLs[shortURL.UID] = shortURL
	m.userShortURLs[shortURL.UserID] = append(m.userShortURLs[shortURL.UserID], shortURL)

	return nil
}

func (m *MemoryRepository) BatchSave(_ context.Context, shortURLs []*models.ShortURL) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shortURL := range shortURLs {
		isFound := false

		userShortURLs, ok := m.userShortURLs[shortURL.UserID]
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

		m.shortURLs[shortURL.UID] = shortURL
		m.userShortURLs[shortURL.UserID] = append(m.userShortURLs[shortURL.UserID], shortURL)
	}

	return nil
}

func (m *MemoryRepository) BatchDelete(_ context.Context, shortURLs []*models.ShortURL) error {
	if len(shortURLs) == 0 {
		return ErrNothingToDelete
	}

	for _, shortURL := range shortURLs {
		shortURL.IsDeleted = true
	}

	return nil
}

func (m *MemoryRepository) FindAllByUserIDAndUIDs(
	_ context.Context,
	userID uuid.UUID,
	uids []models.UID,
) ([]*models.ShortURL, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var shortURLs []*models.ShortURL

	for _, uid := range uids {
		shortURL, ok := m.shortURLs[uid]
		if ok && shortURL.UserID == userID {
			shortURLs = append(shortURLs, shortURL)
		}
	}

	if len(shortURLs) == 0 {
		return nil, ErrNotFound
	}

	return shortURLs, nil
}

func (m *MemoryRepository) Ping(_ context.Context) error {
	return nil
}
