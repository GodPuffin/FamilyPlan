package random

import (
	"math/rand"
	"time"
)

// GenerateJoinCode creates a random plan join code.
func GenerateJoinCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = charset[rng.Intn(len(charset))]
	}

	return string(code)
}

// GenerateToken creates a random auth token.
func GenerateToken() string {
	const (
		tokenLength = 32
		charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	token := make([]byte, tokenLength)
	for i := range token {
		token[i] = charset[rng.Intn(len(charset))]
	}

	return string(token)
}
