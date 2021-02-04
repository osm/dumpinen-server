package main

import (
	"math/rand"
	"time"
)

// charset contains the valid characters that we want to use when we generate
// a random string.
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"

// rndStr returns a random string with the given length.
func rndStr(l int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, l)
	for i := range b {
		b[i] = charset[rnd.Intn(len(charset))]
	}

	return string(b)
}
