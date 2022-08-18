package models

import (
	"fmt"

	"github.com/tmitry/shorturl/internal/app/config"
)

type ShortURL struct {
	id  int
	UID UID
	URL URL
}

func NewShortURL(id int, url URL, uid UID) *ShortURL {
	return &ShortURL{
		id:  id,
		URL: url,
		UID: uid,
	}
}

func (s ShortURL) GetShortURL() string {
	return fmt.Sprintf("%s://%s/%s", config.DefaultProtocol, config.DefaultAddr, s.UID.String())
}
