package random

import (
	crand "crypto/rand"
	"math/big"
)

// GenerateJoinCode creates a random plan join code.
func GenerateJoinCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	code, err := secureString(length, charset)
	if err != nil {
		return ""
	}

	return code
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
