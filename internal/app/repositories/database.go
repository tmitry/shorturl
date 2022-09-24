package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // solely for its side effects (initialization)
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/models"
)

const (
	defaultQueryTimeout = 5
)

type DatabaseRepository struct {
	db *sql.DB
}

func NewDatabaseRepository(databaseCfg *configs.DatabaseConfig) *DatabaseRepository {
	database, err := sql.Open("pgx", databaseCfg.DSN)
	if err != nil {
		log.Panic(err)
	}

	rep := &DatabaseRepository{
		db: database,
	}

	err = rep.CreateDatabase()
	if err != nil {
		log.Panic(err)
	}

	return rep
}

func (d DatabaseRepository) FindOneByUID(ctx context.Context, uid models.UID) (*models.ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout*time.Second)
	defer cancel()

	shortURL := models.NewShortURL(0, "", "", uuid.UUID{})

	err := d.db.QueryRowContext(ctx, "SELECT id, uid, url, user_id FROM short_url WHERE uid = $1", uid).Scan(
		&shortURL.ID,
		&shortURL.UID,
		&shortURL.URL,
		&shortURL.UserID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
	}

	return shortURL, nil
}

func (d DatabaseRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*models.ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout*time.Second)
	defer cancel()

	var userShortURLs []*models.ShortURL

	rows, err := d.db.QueryContext(ctx, "SELECT id, uid, url, user_id FROM short_url WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
	}

	defer rows.Close()

	for rows.Next() {
		shortURL := models.NewShortURL(0, "", "", uuid.UUID{})

		err = rows.Scan(
			&shortURL.ID,
			&shortURL.UID,
			&shortURL.URL,
			&shortURL.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
		}

		userShortURLs = append(userShortURLs, shortURL)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
	}

	if len(userShortURLs) == 0 {
		return nil, ErrNotFound
	}

	return userShortURLs, nil
}

func (d DatabaseRepository) Save(
	ctx context.Context,
	url models.URL,
	userID uuid.UUID,
	hashMinLength int,
	hashSalt string,
) (*models.ShortURL, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout*time.Second)
	defer cancel()

	query := "INSERT INTO short_url(url, user_id) VALUES($1, $2) RETURNING id"

	id := 0
	if err := d.db.QueryRowContext(ctx, query, url, userID).Scan(&id); err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	shortURL := models.NewShortURL(id, url, models.NewUID(id, hashMinLength, hashSalt), userID)

	_, err := d.db.ExecContext(ctx, "UPDATE short_url SET uid = $1 WHERE id = $2", shortURL.UID, shortURL.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	return shortURL, nil
}

func (d DatabaseRepository) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout*time.Second)
	defer cancel()

	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToPing, err)
	}

	return nil
}

func (d DatabaseRepository) CreateDatabase() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultQueryTimeout*time.Second)
	defer cancel()

	query := `
CREATE TABLE IF NOT EXISTS short_url (
	id SERIAL,
	uid VARCHAR(32),
	url TEXT NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	CONSTRAINT short_url_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_uid_idx ON short_url (uid);
CREATE INDEX IF NOT EXISTS short_url_user_id_idx ON short_url (user_id);
`

	if _, err := d.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToCreateDatabase, err)
	}

	return nil
}
