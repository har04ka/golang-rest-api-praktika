package utils

import (
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func ComparePassword(password, hash string) error {
	password = strings.TrimSpace(password)
	hash = strings.TrimSpace(hash)

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
