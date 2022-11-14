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
	shortURL := models.NewShortURL(0, "", "", uuid.UUID{})

	err := d.db.QueryRowContext(
		ctx,
		"SELECT id, uid, url, user_id, is_deleted FROM short_url WHERE uid = $1",
		uid,
	).Scan(
		&shortURL.ID,
		&shortURL.UID,
		&shortURL.URL,
		&shortURL.UserID,
		&shortURL.IsDeleted,
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

	rows, err := d.db.QueryContext(
		ctx,
		"SELECT id, uid, url, user_id, is_deleted FROM short_url WHERE user_id = $1",
		userID,
	)
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
			&shortURL.IsDeleted,
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

func (d DatabaseRepository) Save(ctx context.Context, shortURL *models.ShortURL) (fnErr error) {
	transaction, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(transaction *sql.Tx) {
		err := transaction.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(transaction)

	insertTxStmt, err := transaction.Prepare(`
INSERT INTO short_url(url, uid, user_id) VALUES($1, $2, $3) 
ON CONFLICT(user_id, url) DO NOTHING RETURNING id
`)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(insertTxStmt *sql.Stmt) {
		err := insertTxStmt.Close()
		if err != nil {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(insertTxStmt)

	selectTxStmt, err := transaction.Prepare(`
SELECT id, uid FROM short_url WHERE user_id = $1 AND url = $2
`)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(selectTxStmt *sql.Stmt) {
		err := selectTxStmt.Close()
		if err != nil {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(selectTxStmt)

	row := insertTxStmt.QueryRowContext(ctx, shortURL.URL, shortURL.UID, shortURL.UserID)
	if row.Err() != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, row.Err())
	}

	if err := row.Scan(&shortURL.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := selectTxStmt.QueryRowContext(ctx, shortURL.UserID, shortURL.URL).Scan(
				&shortURL.ID,
				&shortURL.UID,
			)
			if err != nil {
				return fmt.Errorf("%s: %w", messageFailedToFind, err)
			}

			return ErrURLDuplicate
		}

		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	return nil
}

func (d DatabaseRepository) BatchSave(ctx context.Context, shortURLs []*models.ShortURL) (fnErr error) {
	transaction, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(transaction *sql.Tx) {
		err := transaction.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(transaction)

	insertTxStmt, err := transaction.Prepare(`
INSERT INTO short_url(url, uid, user_id) VALUES($1, $2, $3) 
ON CONFLICT(user_id, url) DO NOTHING RETURNING id
`)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(insertTxStmt *sql.Stmt) {
		err := insertTxStmt.Close()
		if err != nil {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(insertTxStmt)

	selectTxStmt, err := transaction.Prepare(`
SELECT id, uid FROM short_url WHERE user_id = $1 AND url = $2
`)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	defer func(selectTxStmt *sql.Stmt) {
		err := selectTxStmt.Close()
		if err != nil {
			fnErr = fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}(selectTxStmt)

	for _, shortURL := range shortURLs {
		row := insertTxStmt.QueryRowContext(ctx, shortURL.URL, shortURL.UID, shortURL.UserID)
		if row.Err() != nil {
			return fmt.Errorf("%s: %w", messageFailedToSave, row.Err())
		}

		if err := row.Scan(&shortURL.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err := selectTxStmt.QueryRowContext(ctx, shortURL.UserID, shortURL.URL).Scan(
					&shortURL.ID,
					&shortURL.UID,
				)
				if err != nil {
					return fmt.Errorf("%s: %w", messageFailedToFind, err)
				}

				continue
			}

			return fmt.Errorf("%s: %w", messageFailedToSave, err)
		}
	}

	if err = transaction.Commit(); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToSave, err)
	}

	return nil
}

func (d DatabaseRepository) BatchDelete(ctx context.Context, shortURLs []*models.ShortURL) error {
	if len(shortURLs) == 0 {
		return ErrNothingToDelete
	}

	ids := make([]int, 0, len(shortURLs))

	for _, shortURL := range shortURLs {
		ids = append(ids, shortURL.ID)
	}

	_, err := d.db.ExecContext(ctx, "UPDATE short_url SET is_deleted = true WHERE id = ANY($1)", ids)
	if err != nil {
		return fmt.Errorf("%s: %w", messageFailedToDelete, err)
	}

	return nil
}

func (d DatabaseRepository) FindAllByUserIDAndUIDs(
	ctx context.Context,
	userID uuid.UUID,
	uids []models.UID,
) (_ []*models.ShortURL, fnErr error) {
	var shortURLs []*models.ShortURL

	rows, err := d.db.QueryContext(
		ctx,
		"SELECT id, uid, url, user_id, is_deleted FROM short_url WHERE user_id = $1 AND uid = ANY($2)",
		userID,
		uids,
	)
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
			&shortURL.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
		}

		shortURLs = append(shortURLs, shortURL)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", messageFailedToFind, err)
	}

	if len(shortURLs) == 0 {
		return nil, ErrNotFound
	}

	return shortURLs, nil
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
	is_deleted BOOL NOT NULL DEFAULT false,
	CONSTRAINT short_url_pkey PRIMARY KEY (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS short_url_uid_idx ON short_url (uid);
CREATE UNIQUE INDEX IF NOT EXISTS short_url_user_id_url_idx ON short_url (user_id, url);
`

	if _, err := d.db.Exec(query); err != nil {
		return fmt.Errorf("%s: %w", messageFailedToCreateDatabase, err)
	}

	return nil
}
