package utils

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/speps/go-hashids/v2"
	"github.com/tmitry/shorturl/internal/app/models"
)

type UIDGenerator interface {
	Generate() (models.UID, error)
	GetPattern() string
	IsValid(u models.UID) (bool, error)
}

type HashidsUIDGenerator struct {
	hashMinLength int
	hashSalt      string
	encoder       *hashids.HashID
	mu            sync.Mutex
}

func NewHashidsUIDGenerator(hashMinLength int, hashSalt string) *HashidsUIDGenerator {
	hd := hashids.NewData()
	hd.Salt = hashSalt
	hd.MinLength = hashMinLength

	encoder, err := hashids.NewWithData(hd)
	if err != nil {
		log.Panic(err)
	}

	return &HashidsUIDGenerator{
		hashMinLength: hashMinLength,
		hashSalt:      hashSalt,
		encoder:       encoder,
	}
}

func (gen *HashidsUIDGenerator) Generate() (models.UID, error) {
	gen.mu.Lock()
	uniqueInt := time.Now().UnixMilli()
	time.Sleep(time.Millisecond)
	gen.mu.Unlock()

	str, err := gen.encoder.EncodeInt64([]int64{uniqueInt})
	if err != nil {
		return "", fmt.Errorf("failed to encode: %w", err)
	}

	return models.UID(str), nil
}

func (gen *HashidsUIDGenerator) GetPattern() string {
	return fmt.Sprintf("[0-9a-zA-Z]{%d,}", gen.hashMinLength)
}

func (gen *HashidsUIDGenerator) IsValid(u models.UID) (bool, error) {
	matched, err := regexp.MatchString(gen.GetPattern(), u.String())
	if err != nil {
		return false, fmt.Errorf("failed to match string: %w", err)
	}

	return matched, nil
}
