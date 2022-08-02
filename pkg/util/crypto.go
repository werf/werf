package util

import "crypto/rand"

func GenerateConsistentRandomString(n int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := generateRandomBytes(n)
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err == nil {
		return b
	} else {
		panic(err)
	}
}
