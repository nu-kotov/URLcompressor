package utils

import (
	"fmt"
	"hash/fnv"

	"github.com/sqids/sqids-go"
)

// Hash вычисляет хеш строки байт.
func Hash(bytes []byte) uint64 {

	h := fnv.New64a()
	_, _ = h.Write(bytes)

	return h.Sum64()
}

func HashOriginalURL(originalURL []byte) (string, error) {
	sqids, err := sqids.New()
	if err != nil {
		return "", fmt.Errorf("sqids lib error: %w", err)
	}

	bodyHash := Hash(originalURL)
	shortID, err := sqids.Encode([]uint64{bodyHash})
	if err != nil {
		return "", fmt.Errorf("short ID creating error: %w", err)
	}

	return shortID, nil
}
