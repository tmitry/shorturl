package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/stdlib" // solely for its side effects (initialization)
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/models"
)

type DatabaseRepository struct {
	DB *sql.DB
}

func NewDatabaseRepository(databaseCfg *configs.DatabaseConfig) *DatabaseRepository {
	database, err := sql.Open("pgx", databaseCfg.DSN)
	if err != nil {
		log.Panic(err)
	}

	return &DatabaseRepository{
		DB: database,
	}
}

func (d DatabaseRepository) ReserveID() int {
	// TODO implement me
	panic("implement me")
}

func (d DatabaseRepository) FindOneByUID(uid models.UID) *models.ShortURL {
	// TODO implement me
	panic("implement me")
}

func (d DatabaseRepository) FindAllByUserID(uuid uuid.UUID) []*models.ShortURL {
	// TODO implement me
	panic("implement me")
}

func (d DatabaseRepository) Save(shortURL *models.ShortURL) {
	// TODO implement me
	panic("implement me")
}

func (d DatabaseRepository) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := d.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}
