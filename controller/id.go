package controller

import (
	"crypto/rand"
	"math/big"
)

var (
	idChars = "abcdef0123456789"
)

// NewID creates a new ID.
func NewID() string {
	b := make([]byte, 3)

	for i := range b {
		max := big.NewInt(int64(len(idChars)))
		j, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic(err)
		}

		b[i] = idChars[int(j.Int64())]
	}

	return string(b)
}
