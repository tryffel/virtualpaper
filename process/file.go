package process

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
	"tryffel.net/go/virtualpaper/storage"
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

type fileProcessor struct {
	*Task
	input   chan fileOp
	file    string
	rawFile *os.File
}

func newFileProcessor(id int, db *storage.Database) *fileProcessor {
	fp := &fileProcessor{
		Task:  newTask(id, db),
		input: make(chan fileOp, 5),
	}
	fp.idle = true
	fp.runFunc = fp.waitFile
	return fp
}

func (fp *fileProcessor) waitFile() {
	timer := time.NewTimer(time.Millisecond * 10)
	select {
	case <-timer.C:
		// pass

	case fileOp := <-fp.input:
		fp.file = fileOp.file
		fp.processFile()
	}
}

func (fp *fileProcessor) processFile() {
	logrus.Infof("task %d, process file %s", fp.id, fp.file)

	fp.lock.Lock()
	fp.idle = false
	fp.lock.Unlock()
	var err error

	fp.rawFile, err = os.Open(fp.file)

	defer fp.cleanup()

	if err != nil {
		logrus.Errorf("process file %s, open: %v", fp.file, err)
		return
	}

	duplicate, err := fp.isDuplicate()
	if duplicate {
		logrus.Infof("file %s is a duplicate, ignore file", fp.file)
		return
	}

	if err != nil {
		logrus.Errorf("get duplicate status: %v", err)
		return
	}

}

func (fp *fileProcessor) cleanup() {
	logrus.Infof("Stop processing file %s", fp.file)
	fp.rawFile.Close()
	fp.file = ""
	fp.lock.Lock()
	fp.idle = true
	fp.lock.Unlock()
}

func (fp *fileProcessor) isDuplicate() (bool, error) {
	hash, err := getHash(fp.rawFile)
	if err != nil {
		return false, err
	}

	document, err := fp.db.DocumentStore.GetByHash(hash)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if document != nil {
		return true, nil
	}
	return false, nil
}
