package process

import (
	"os"
	"path"
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
		t.Log("output is empty: ", file == nil)

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
