package utils

import (
	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = bcrypt.DefaultCost

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
