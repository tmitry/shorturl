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

func (s ShortURL) GetShortURL() URL {
	return URL(fmt.Sprintf("%s/%s", config.ServerCfg.BaseURL, s.UID.String()))
}
