package models

import (
	"fmt"
	"regexp"

	"github.com/speps/go-hashids/v2"
)

type UID string

func NewUID(identifier, hashMinLength int, hashSalt string) UID {
	hd := hashids.NewData()
	hd.Salt = hashSalt
	hd.MinLength = hashMinLength
	h, _ := hashids.NewWithData(hd)

	str, _ := h.Encode([]int{identifier})

	return UID(str)
}

func GetPatternUID(hashMinLength int) string {
	return fmt.Sprintf("[0-9a-zA-Z]{%d,}", hashMinLength)
}

func (u UID) IsValid(hashMinLength int) (bool, error) {
	matched, err := regexp.MatchString(GetPatternUID(hashMinLength), u.String())
	if err != nil {
		return false, fmt.Errorf("failed to match string: %w", err)
	}

	return matched, nil
}

func (u UID) String() string {
	return string(u)
}
