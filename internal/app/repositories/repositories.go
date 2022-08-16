package repositories

import "github.com/tmitry/shorturl/internal/app/models"

type Repository interface {
	Find(u models.UID) *models.ShortURL
	Add(u models.URL) *models.ShortURL
}
