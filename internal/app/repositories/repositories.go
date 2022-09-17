package repositories

import (
	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

type Repository interface {
	ReserveID() int
	FindOneByUID(uid models.UID) *models.ShortURL
	FindAllByUserID(uuid uuid.UUID) []*models.ShortURL
	Save(shortURL *models.ShortURL)
}
