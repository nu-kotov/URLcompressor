package utils

import (
	"hash/fnv"
)

// Hash вычисляет хеш строки байт.
func Hash(bytes []byte) uint64 {

	h := fnv.New64a()
	_, _ = h.Write(bytes)

	return h.Sum64()
}
