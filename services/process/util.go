package process

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/storage"
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

// DeleteDocument deletes original document and its preview file.
func DeleteDocument(docId string) error {
	previewPath := storage.PreviewPath(docId)
	docPath := storage.DocumentPath(docId)

	logrus.Debugf("delete preview file %s", previewPath)
	err := os.Remove(previewPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
			logrus.Warnf("document %s preview file not found, continuing delete operation...", docId)
		} else {
			return fmt.Errorf("remove thumbnail: %v", err)
		}
	}
	logrus.Debugf("delete document file %s", previewPath)
	err = os.Remove(docPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
			logrus.Warnf("document %s file not found, continuing delete operation...", docId)
		} else {
			return fmt.Errorf("remove document file: %v", err)
		}
	}
	return nil
}
