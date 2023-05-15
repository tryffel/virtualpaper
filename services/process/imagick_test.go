//go:build test_integration
// +build test_integration

package process

import (
	"os"
	"path"
	"path/filepath"
	"testing"
	"tryffel.net/go/virtualpaper/config"
)

func TestGetImagickVersion(t *testing.T) {
	config.ConfigFromViper()
	config.InitConfig()
	ver := GetImagickVersion()
	if ver != "" {
		t.Errorf("imagick returned version: %s, want: ''", ver)
	}

	config.C.Processing.ImagickBin = "/usr/local/bin/convert"
	ver = GetImagickVersion()
	if ver == "" {
		t.Errorf("imagick returned version: %s, want: non-empty", ver)
	}
}

func TestGenerateThumbnail(t *testing.T) {
	config.ConfigFromViper()
	config.InitConfig()
	config.C.Processing.ImagickBin = "/usr/local/bin/convert"

	wd, _ := os.Getwd()
	wd = path.Dir(wd)

	destDir := t.TempDir()
	inputDir := "e2e/test_data"

	testFile := func(inputName string, outputName string) {
		err := generateThumbnail(path.Join(wd, inputDir, inputName), path.Join(destDir, outputName), 0, 500, "png")
		if err != nil {
			t.Errorf("imagick generate thumbnail: %v", err)
		}

		output := path.Join(destDir, outputName)
		file, err := os.Open(output)
		if err != nil {
			if err == os.ErrNotExist {
				t.Errorf("output file does not exist")
			} else {
				t.Error("cannot open output file: ", err)
			}
			t.Fail()
		}
		file.Close()
	}

	testFile("jpg-1.jpg", "jpg-1.png")
	testFile("pdf-1.pdf", "pdf-1.png")
	testFile("png-1.png", "png-1.png")
}

func TestGeneratePicture(t *testing.T) {
	config.ConfigFromViper()
	config.InitConfig()
	config.C.Processing.ImagickBin = "/usr/local/bin/convert"

	wd, _ := os.Getwd()
	wd = path.Dir(wd)

	destDir := t.TempDir()
	inputDir := "e2e/test_data"
	inputFile := "pdf-1.pdf"
	outputFile := "pdf-1.png"
	pages := 2

	err := generatePicture(path.Join(wd, inputDir, inputFile), path.Join(destDir, outputFile))
	if err != nil {
		t.Errorf("generate picture: %v", err)
	}

	foundFiles, err := filepath.Glob(path.Join(destDir, "pdf-1*"))
	if len(foundFiles) != pages {
		t.Errorf("output files count does not match, got %d, want: %d", len(foundFiles), pages)
	}
}
