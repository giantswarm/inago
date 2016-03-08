package controller

import (
	"math/rand"
)

var (
	idChars = "abcdef0123456789"
)

// NewID creates a new ID.
func NewID() string {
	b := make([]byte, 3)
	ns := []int{}

	for i := range b {
		if i%len(idChars) == 0 {
			ns = rand.Perm(len(idChars))
		}

		b[i] = idChars[ns[i]]
	}

	return string(b)
}
