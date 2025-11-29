package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hash string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func HashTokenHMAC(token string) string {
	var SecretKey = []byte(os.Getenv("SECRET_KEY"))

	mac := hmac.New(sha256.New, SecretKey)
	mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil))
}

func CompareTokenHash(token, hash string) bool {
	secretKey := []byte(os.Getenv("SECRET_KEY"))

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(token))
	expectedMAC := mac.Sum(nil)

	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return false
	}

	return hmac.Equal(expectedMAC, hashBytes)
}
