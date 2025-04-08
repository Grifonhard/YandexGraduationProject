package utils

import (
	"crypto/rand"
	"math/big"
)

const GEN_SYMBOLS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// GenerateString генерирует стрингу из букв и цифр длинной length
func GenerateString(length int) (string, error) {
	password := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(GEN_SYMBOLS))))
		if err != nil {
			return "", err
		}
		password[i] = GEN_SYMBOLS[n.Int64()]
	}

	return string(password), nil
}