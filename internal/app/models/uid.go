package models

import (
	"fmt"
	"regexp"

	"github.com/tmitry/shorturl/internal/app/config"
)

type UID string

func (s UID) IsValid() (bool, error) {
	matched, err := regexp.MatchString(config.PatternUID, s.String())
	if err != nil {
		return false, fmt.Errorf("failed to match string: %w", err)
	}

	return matched, nil
}

func (s UID) String() string {
	return string(s)
}
