//go:build test_integration
// +build test_integration

package process

import (
	"os"
	"path"
	"strings"
	"testing"

	"tryffel.net/go/virtualpaper/config"
)

func Test_runOcr(t *testing.T) {
	t.Log("Test OCR")
	config.ConfigFromViper()
	config.InitConfig()
	config.C.Processing.ImagickBin = "/usr/local/bin/convert"

	if GetTesseractVersion() == "" {
		t.Log("no tesseract found, skipping test")
		t.Skip()
		return
	}

	wd, _ := os.Getwd()
	wd = path.Dir(wd)

	inputDir := "e2e/test_data"

	t.Log("Extract contents from JPG")
	text, err := runOcr(path.Join(wd, inputDir, "jpg-1.jpg"), "test")
	if err != nil {
		t.Errorf("run ocr for jpg: %v", err)
	}
	if !strings.HasPrefix(text, "Lorem ipsum") || len(text) != 4003 {
		t.Error("jpg text doesn't match")
	}

	t.Log("Extract contents from PNG")
	text, err = runOcr(path.Join(wd, inputDir, "png-1.png"), "test")
	if err != nil {
		t.Errorf("run ocr for png: %v", err)
	}
	if !strings.HasPrefix(text, "Lorem ipsum") || len(text) != 4003 {
		t.Error("png text doesn't match")
	}

	t.Log("Extract contents from PDF")
	text, err = runOcr(path.Join(wd, inputDir, "pdf-1.pdf"), "test")
	if err != nil {
		t.Errorf("run ocr for pdf: %v", err)
	}

	// pdf extraction is not deterministic and text length may vary
	if !strings.HasPrefix(text, "Lorem ipsum") || len(text) < 3000 || len(text) > 6000 {
		t.Error("pdf text doesn't match")
	}
}
