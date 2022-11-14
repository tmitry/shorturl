package utils

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/configs"
	"github.com/tmitry/shorturl/internal/app/models"
	"github.com/tmitry/shorturl/internal/app/repositories"
)

type DeletionBuffer interface {
	Push(uids []models.UID, userID uuid.UUID)
}

type BackgroundDeletionBuffer struct {
	rep                repositories.Repository
	bufferMaxSize      int
	bufferClearTimeout time.Duration
	shortURLs          []*models.ShortURL
	channel            chan *models.ShortURL
	uidGenerator       UIDGenerator
}

func NewBackgroundDeletionBuffer(
	rep repositories.Repository,
	appCfg *configs.AppConfig,
	uidGenerator UIDGenerator,
) *BackgroundDeletionBuffer {
	buf := &BackgroundDeletionBuffer{
		rep:                rep,
		bufferMaxSize:      appCfg.DeletionBufferMaxSize,
		bufferClearTimeout: time.Duration(appCfg.DeletionBufferClearTimeout) * time.Second,
		shortURLs:          []*models.ShortURL{},
		channel:            make(chan *models.ShortURL),
		uidGenerator:       uidGenerator,
	}

	buf.newWorker()

	return buf
}

func (buf *BackgroundDeletionBuffer) Push(uids []models.UID, userID uuid.UUID) {
	go func() {
		for _, uid := range uids {
			isValid, err := buf.uidGenerator.IsValid(uid)
			if err != nil || !isValid {
				return
			}
		}

		shortURLs, err := buf.rep.FindAllByUserIDAndUIDs(context.Background(), userID, uids)
		if err != nil {
			return
		}

		for _, shortURL := range shortURLs {
			buf.channel <- shortURL
		}
	}()
}

func (buf *BackgroundDeletionBuffer) newWorker() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				buf.newWorker()
				log.Println(r)
			}
		}()

		ticker := time.NewTicker(buf.bufferClearTimeout)

		for {
			select {
			case shortURL := <-buf.channel:
				buf.shortURLs = append(buf.shortURLs, shortURL)
				if len(buf.shortURLs) == buf.bufferMaxSize {
					buf.flush()
				}

				ticker.Reset(buf.bufferClearTimeout)
			case <-ticker.C:
				buf.flush()
			}
		}
	}()
}

func (buf *BackgroundDeletionBuffer) flush() {
	err := buf.rep.BatchDelete(context.Background(), buf.shortURLs)
	if err != nil {
		if !errors.Is(err, repositories.ErrNothingToDelete) {
			log.Println(err.Error())
		}

		return
	}

	buf.shortURLs = nil
}
