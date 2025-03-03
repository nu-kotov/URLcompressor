package utils

import (
	"hash/fnv"
)

func Hash(bytes []byte) uint64 {

	h := fnv.New64a()
	_, _ = h.Write(bytes)

	return h.Sum64()
}
