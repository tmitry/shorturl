package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/models"
)

const (
	messageFailedToSave           = "failed to save"
	messageFailedToFind           = "failed to find"
	messageFailedToPing           = "failed to ping"
	messageFailedToCreateDatabase = "failed to create database"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrURLDuplicate = errors.New("duplicate url")
)

type Repository interface {
	FindOneByUID(ctx context.Context, uid models.UID) (*models.ShortURL, error)

	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*models.ShortURL, error)

	Save(ctx context.Context, shortURL *models.ShortURL) error

	BatchSave(ctx context.Context, shortURLs []*models.ShortURL) error

	Ping(ctx context.Context) error
}
