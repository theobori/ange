package common

import (
	"crypto/rand"
	"fmt"
)

func GenerateToken(bytesAmount int) string {
	if bytesAmount < 0 {
		bytesAmount = 1
	}

	b := make([]byte, bytesAmount)
	rand.Read(b)

	return fmt.Sprintf("%x", b)
}
