package process

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

func (fp *fileProcessor) generateThumbnail() error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessThumbnail,
		CreatedAt:  time.Now(),
	}

	err := fp.ensureFileOpen()
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "generate thumbnail")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}
	defer fp.completeProcessingStep(process, job)

	output := storage.PreviewPath(fp.document.Id)
	err = storage.CreatePreviewDir(fp.document.Id)
	if err != nil {
		return fmt.Errorf("create thumbnail output dir: %v", err)
	}

	name := fp.rawFile.Name()
	err = generateThumbnail(name, output, 0, 500, fp.document.Mimetype)
	if err != nil {
		job.Status = models.JobFailure
		job.Message += "; " + err.Error()
		return fmt.Errorf("call imagick: %v", err)
	}

	job.Status = models.JobFinished
	return nil
}

func generateThumbnailPlainText(rawFile string, previewFile string, size int) error {
	logrus.Debugf("generate thumbnail for text file")

	inputFile, err := os.Open(rawFile)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(previewFile)
	if err != nil {
		return fmt.Errorf("create output file: %v", err)
	}
	defer outputFile.Close()

	// A4 sized preview
	height := size
	width := int(float64(size) * 0.707)
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	y := 20

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.White)
		}
	}
	face := basicfont.Face7x13

	// split text to
	splitTextLines := func(text string) []string {
		maxY := 46

		if len(text) < maxY {
			return []string{text}
		}

		textLeft := text
		lines := make([]string, 0, 2)

		for true {
			if len(textLeft) == 0 {
				break
			}
			if len(textLeft) < maxY {
				lines = append(lines, textLeft)
				break
			}

			// find last whitespace to split at

			for i := maxY; i > 0; i-- {
				if textLeft[i] == ' ' {
					line := textLeft[0:i]
					line = strings.Trim(line, " \n")
					textLeft = textLeft[i+1:]
					lines = append(lines, line)
					break
				}
			}
		}

		return lines
	}

	// print text to file. Return true if print successful.
	// When page is full, return false.
	addText := func(text string) bool {
		splits := splitTextLines(text)

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}),
			Face: face,
		}

		for _, row := range splits {
			// page full
			if y > height-20 {
				return false
			}

			d.Dot = fixed.Point26_6{fixed.Int26_6(4 * 64), fixed.Int26_6(y * 64)}
			d.DrawString(row)
			y += 20
		}

		return true
	}

	maxRows := 24
	maxChars := 1100

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	text := ""
	totalRows := 0

	for scanner.Scan() {
		row := scanner.Text()
		if text != "" {
			text += " "
		}
		totalRows += 1
		text += row
		if len(text) > maxChars || totalRows > maxRows {
			break
		}

		written := addText(row)
		if !written {
			// page full
			break
		}

	}

	err = png.Encode(outputFile, img)
	if err != nil {
		return fmt.Errorf("flush output buffer: %v", err)
	}
	return nil
}

func (fp *fileProcessor) updateThumbnail(doc *models.Document, file *os.File) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessThumbnail,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "generate thumbnail")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}
	job.Message = "Generate thumbnail"
	defer fp.completeProcessingStep(process, job)

	output := storage.PreviewPath(fp.document.Id)

	err = storage.CreatePreviewDir(fp.document.Id)
	if err != nil {
		logrus.Errorf("create preview dir: %v", err)
	}

	logrus.Infof("generate thumbnail for document %s", fp.document.Id)
	err = generateThumbnail(file.Name(), output, 0, 500, process.Document.Mimetype)

	err = fp.db.DocumentStore.Update(storage.UserIdInternal, doc)
	if err != nil {
		logrus.Errorf("update document record: %v", err)
	}

	if err != nil {
		job.Status = models.JobFailure
		job.Message += "; " + err.Error()
		return fmt.Errorf("call imagick: %v", err)
	}
	job.Status = models.JobFinished
	return nil
}
