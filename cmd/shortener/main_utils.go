package main

import (
	"hash/fnv"
)

func hash(bytes []byte) uint64 {

	h := fnv.New64a()
	h.Write([]byte(bytes))

	return h.Sum64()
}
