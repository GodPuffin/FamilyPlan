package random

import (
	crand "crypto/rand"
	"math/big"
	mathrand "math/rand"
	"time"
)

// GenerateJoinCode creates a random plan join code.
func GenerateJoinCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	source := mathrand.NewSource(time.Now().UnixNano())
	rng := mathrand.New(source)

	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = charset[rng.Intn(len(charset))]
	}

	return string(code)
}

// GenerateToken creates a random auth token.
func GenerateToken() (string, error) {
	const (
		tokenLength = 32
		charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	return secureString(tokenLength, charset)
}

func secureString(length int, charset string) (string, error) {
	value := make([]byte, length)
	max := big.NewInt(int64(len(charset)))

	for i := range value {
		index, err := crand.Int(crand.Reader, max)
		if err != nil {
			return "", err
		}

		value[i] = charset[index.Int64()]
	}

	return string(value), nil
}
