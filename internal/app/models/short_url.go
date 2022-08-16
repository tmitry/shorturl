package models

import (
	"fmt"

	"github.com/speps/go-hashids/v2"
	"github.com/tmitry/shorturl/internal/app/config"
)

type ShortURL struct {
	id  int
	UID UID
	URL URL
}

func NewShortURL(id int, url URL) *ShortURL {
	uid := UID(generateUniqueHash(id))

	return &ShortURL{
		id:  id,
		URL: url,
		UID: uid,
	}
}

func (s ShortURL) GetShortURL() string {
	return fmt.Sprintf("%s://%s/%s", config.DefaultProtocol, config.DefaultAddr, s.UID.String())
}

func generateUniqueHash(identifier int) string {
	hd := hashids.NewData()
	hd.Salt = config.HashSalt
	hd.MinLength = config.HashMinLength
	h, _ := hashids.NewWithData(hd)

	uid, _ := h.Encode([]int{identifier})

	return uid
}
