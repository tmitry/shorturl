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

var ErrNotFound = errors.New("not found")

type Repository interface {
	FindOneByUID(requestCtx context.Context, uid models.UID) (*models.ShortURL, error)

	FindAllByUserID(requestCtx context.Context, userID uuid.UUID) ([]*models.ShortURL, error)

	Save(
		requestCtx context.Context,
		url models.URL,
		userID uuid.UUID,
		hashMinLength int,
		hashSalt string,
	) (*models.ShortURL, error)

	Ping(requestCtx context.Context) error
}
