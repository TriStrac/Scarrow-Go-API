package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

func GenerateDeviceID(prefix string) string {
	return fmt.Sprintf("%s-%s-%s", prefix, generateRandomString(4), generateRandomString(4))
}

func GenerateSecret(length int) string {
	return generateRandomString(length)
}
