// Package auth provides password hashing, checking, and random password generation
// utilities for authenticating admin users within the server manager.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plaintext password using bcrypt and returns the hashed string.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}

// CheckPassword verifies whether the provided plaintext password matches the given bcrypt hash.
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// GenerateRandomPassword creates a random password of the specified byte length,
// returning both the plaintext and its hashed version.
func GenerateRandomPassword(length int) (plain string, hashed string, err error) {
	raw := make([]byte, length)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	plain = base64.RawURLEncoding.EncodeToString(raw)
	hashed, err = HashPassword(plain)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash password: %w", err)
	}
	return plain, hashed, nil
}
