package repositories

import "github.com/tmitry/shorturl/internal/app/models"

type Repository interface {
	ReserveID() int
	Find(uid models.UID) *models.ShortURL
	Add(shortURL *models.ShortURL)
}
