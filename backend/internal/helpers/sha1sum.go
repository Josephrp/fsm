package helpers

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

func CalculateSHA1(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha1.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	sum := hasher.Sum(nil)
	return hex.EncodeToString(sum), nil
}
