package models

import (
	"fmt"

	"github.com/google/uuid"
)

type ShortURL struct {
	ID        int
	UID       UID
	URL       URL
	UserID    uuid.UUID
	IsDeleted bool
}

func NewShortURL(id int, url URL, uid UID, userID uuid.UUID) *ShortURL {
	return &ShortURL{
		ID:        id,
		UID:       uid,
		URL:       url,
		UserID:    userID,
		IsDeleted: false,
	}
}

func (s ShortURL) GetShortURL(baseURL string) URL {
	return URL(fmt.Sprintf("%s/%s", baseURL, s.UID))
}
