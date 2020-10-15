package process

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// get md5 hash for file
func getHash(file *os.File) (string, error) {
	hash := md5.New()

	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	b := hash.Sum(nil)[:16]
	return hex.EncodeToString(b), nil
}
