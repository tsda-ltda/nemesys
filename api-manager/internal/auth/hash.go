package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// Hash returns a hash of cost equal to 14.
func Hash(secret string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(secret), cost)
	return string(bytes), err
}

// CheckHash compare a secret and a hash and return the check result.
func CheckHash(secret, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(secret))
	return err == nil
}
