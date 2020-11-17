package process

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// GetFileHash returns unique hash for file. It uses md5 for hashing.
func GetFileHash(file *os.File) (string, error) {
	hash := md5.New()

	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	b := hash.Sum(nil)[:16]
	return hex.EncodeToString(b), nil
}

// GetHash returns hash for file by its name.
func GetHash(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", fmt.Errorf("open file: %v", err)
	}

	defer file.Close()
	hash, err := GetFileHash(file)
	return hash, err
}
