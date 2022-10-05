package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // solely for its side effects (initialization)
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/models"
)

type DatabaseRepository struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	updateStmt *sql.Stmt
}

func NewDatabaseRepository(databaseCfg *configs.DatabaseConfig) *DatabaseRepository {
	database, err := sql.Open("pgx", databaseCfg.DSN)
	if err != nil {
		log.Panic(err)
	}

	rep := &DatabaseRepository{
		db:         database,
		insertStmt: nil,
		updateStmt: nil,
	}

	err = rep.CreateDatabase()
	if err != nil {
		log.Panic(err)
	}

	insertStmt, err := database.Prepare(
		"INSERT INTO short_url(url, uid, user_id) VALUES($1, '', $2) RETURNING id",
	)
	if err != nil {
		log.Panic(err)
	}

	rep.insertStmt = insertStmt

	updateStmt, err := database.Prepare(
		"UPDATE short_url SET uid = $1 WHERE id = $2",
	)
	if err != nil {
		log.Panic(err)
	}

	rep.updateStmt = updateStmt

	return rep
}

func (d DatabaseRepository) FindOneByUID(ctx context.Context, uid models.UID) (*models.ShortURL, error) {
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

func (d DatabaseRepository) FindAllByUserID(ctx context.Context, userID uuid.UUID) (_ []*models.ShortURL, fnErr error) {
	var userShortURLs []*models.ShortURL

	rows, err := d.db.QueryContext(ctx, "SELECT id, uid, url, user_id FROM short_url WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fnErr = err
		}
	}(rows)

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
) (_ *models.ShortURL, fnErr error) {
	transaction, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(transaction *sql.Tx) {
		err := transaction.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(transaction)

	identifier := 0

	insertTxStmt := transaction.StmtContext(ctx, d.insertStmt)

	row := insertTxStmt.QueryRowContext(ctx, url, userID)
	if row.Err() != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, row.Err())
	}

	if err := row.Scan(&identifier); err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	shortURL := models.NewShortURL(identifier, url, models.NewUID(identifier, hashMinLength, hashSalt), userID)

	updateTxStmt := transaction.StmtContext(ctx, d.updateStmt)

	_, err = updateTxStmt.ExecContext(ctx, shortURL.UID, shortURL.ID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	if err = transaction.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	return shortURL, nil
}

func (d DatabaseRepository) BatchSave(
	ctx context.Context,
	urls []models.URL,
	userID uuid.UUID,
	hashMinLength int,
	hashSalt string,
) (_ []*models.ShortURL, fnErr error) {
	transaction, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(transaction *sql.Tx) {
		err := transaction.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(transaction)

	batchShortURLs := make([]*models.ShortURL, 0, len(urls))

	for _, url := range urls {
		identifier := 0

		insertTxStmt := transaction.StmtContext(ctx, d.insertStmt)

		row := insertTxStmt.QueryRowContext(ctx, url, userID)
		if row.Err() != nil {
			return nil, fmt.Errorf("%s: %w", messageFailedToSave, row.Err())
		}

		if err := row.Scan(&identifier); err != nil {
			return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
		}

		shortURL := models.NewShortURL(identifier, url, models.NewUID(identifier, hashMinLength, hashSalt), userID)

		updateTxStmt := transaction.StmtContext(ctx, d.updateStmt)

		_, err = updateTxStmt.ExecContext(ctx, shortURL.UID, shortURL.ID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
		}

		batchShortURLs = append(batchShortURLs, shortURL)
	}

	if err = transaction.Commit(); err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	return batchShortURLs, nil
}

func (d DatabaseRepository) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToPing, err)
	}

	return nil
}

func (d DatabaseRepository) CreateDatabase() error {
	query := `
CREATE TABLE IF NOT EXISTS short_url (
	id SERIAL,
	uid VARCHAR(32) NOT NULL,
	url TEXT NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	CONSTRAINT short_url_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_uid_idx ON short_url (uid);
CREATE INDEX IF NOT EXISTS short_url_user_id_idx ON short_url (user_id);
`

	if _, err := d.db.Exec(query); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToCreateDatabase, err)
	}

	return nil
}
