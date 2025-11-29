package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (hash string, err error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func HashSHA256() {

}
