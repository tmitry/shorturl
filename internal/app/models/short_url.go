package models

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tmitry/shorturl/internal/app/configs"
)

type ShortURL struct {
	id     int
	UID    UID
	URL    URL
	UserID uuid.UUID
}

func NewShortURL(id int, url URL, uid UID, userID uuid.UUID) *ShortURL {
	return &ShortURL{
		id:     id,
		UID:    uid,
		URL:    url,
		UserID: userID,
	}
}

func (s ShortURL) GetShortURL() URL {
	return URL(fmt.Sprintf("%s/%s", configs.ServerCfg.BaseURL, s.UID.String()))
}
