package utils

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword converts a plain text password into a hashed version
func HashPassword(password string) (string, error) {
	// Cost factor of 12 provides a good balance between security and performance
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a password against a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks password complexity requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// Add more password validation rules as needed
	return nil
}
