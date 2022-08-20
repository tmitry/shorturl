package util

import (
	"github.com/speps/go-hashids/v2"
	"github.com/tmitry/shorturl/internal/app/config"
)

func GenerateUniqueHash(identifier int) string {
	hd := hashids.NewData()
	hd.Salt = config.HashSalt
	hd.MinLength = config.HashMinLength
	h, _ := hashids.NewWithData(hd)

	uid, _ := h.Encode([]int{identifier})

	return uid
}
