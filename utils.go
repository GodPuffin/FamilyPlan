package main

import (
	"math/rand"
	"time"
)

// Generate a random join code for family plans
func generateJoinCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create a new random source
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// Generate the code
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		code[i] = charset[r.Intn(len(charset))]
	}

	return string(code)
}

// Generate a random token for authentication
func generateRandomToken() string {
	const tokenLength = 32
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Create a new random source
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	// Generate the token
	token := make([]byte, tokenLength)
	for i := range token {
		token[i] = charset[r.Intn(len(charset))]
	}

	return string(token)
}
