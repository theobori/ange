package hash

import (
	"crypto/sha256"
)

func Generate(s string) string {
	bytes := sha256.Sum256([]byte(s))

	return string(bytes[:])
}

type Hash struct {
	value string
}

func NewHash(s string) *Hash {
	hashString := Generate(s)

	return &Hash{
		value: hashString,
	}
}

func (h *Hash) Verify(s string) bool {
	return h.value == Generate(s)
}

func (h *Hash) String() string {
	return h.value
}
