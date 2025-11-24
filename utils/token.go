package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSessionToken() (string, error) {
	b := make([]byte, 32) // 256 бит
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
