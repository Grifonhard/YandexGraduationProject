package auth

import (
	"crypto/rand"
	"math/big"
)

const PW_SYMBOLS = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

// passwordGenerate генерирует пароль
func passwordGenerate(passwordLength int) (string, error) {
	password := make([]byte, passwordLength)

	for i := 0; i < passwordLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(PW_SYMBOLS))))
		if err != nil {
			return "", err
		}
		password[i] = PW_SYMBOLS[n.Int64()]
	}

	return string(password), nil
}