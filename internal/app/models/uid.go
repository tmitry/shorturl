package models

import (
	"fmt"
	"regexp"

	"github.com/speps/go-hashids/v2"
	"github.com/tmitry/shorturl/internal/app/configs"
)

type UID string

func GenerateUID(identifier int) UID {
	hd := hashids.NewData()
	hd.Salt = configs.AppCfg.HashSalt
	hd.MinLength = configs.AppCfg.HashMinLength
	h, _ := hashids.NewWithData(hd)

	str, _ := h.Encode([]int{identifier})

	return UID(str)
}

func GetPatternUID() string {
	return fmt.Sprintf("[0-9a-zA-Z]{%d,}", configs.AppCfg.HashMinLength)
}

func (u UID) IsValid() (bool, error) {
	matched, err := regexp.MatchString(GetPatternUID(), u.String())
	if err != nil {
		return false, fmt.Errorf("failed to match string: %w", err)
	}

	return matched, nil
}

func (u UID) String() string {
	return string(u)
}
