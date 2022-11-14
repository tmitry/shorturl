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
	messageFailedToDelete         = "failed to delete"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrURLDuplicate    = errors.New("duplicate url")
	ErrNothingToDelete = errors.New("nothing to delete")
)

type Repository interface {
	FindOneByUID(ctx context.Context, uid models.UID) (*models.ShortURL, error)

	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*models.ShortURL, error)

	Save(ctx context.Context, shortURL *models.ShortURL) error

	BatchSave(ctx context.Context, shortURLs []*models.ShortURL) error

	BatchDelete(ctx context.Context, shortURLs []*models.ShortURL) error

	FindAllByUserIDAndUIDs(ctx context.Context, userID uuid.UUID, uids []models.UID) ([]*models.ShortURL, error)

	Ping(ctx context.Context) error
}
