package utils

import (
	"hash/fnv"
	"log"
	"strconv"
)

var (
	hash = fnv.New32a()
)

func StringToUintHash(commentID string) uint {
	hash.Reset()
	hash.Write([]byte(commentID))

	return uint(hash.Sum32())
}

func MustIntToUint(s string) uint {
	id_, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}

	return uint(id_)
}
