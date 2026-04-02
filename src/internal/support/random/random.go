package random

import (
	crand "crypto/rand"
	"math/big"

	"github.com/google/uuid"
)

// GenerateJoinCode creates a random plan join code.
func GenerateJoinCode(length int) (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	return secureString(length, charset)
}

// GenerateToken creates a random auth token.
func GenerateToken() (string, error) {
	const (
		tokenLength = 32
		charset     = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	)

	return secureString(tokenLength, charset)
}

// GenerateUUID creates a random UUID string.
func GenerateUUID() (string, error) {
	return uuid.NewString(), nil
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
