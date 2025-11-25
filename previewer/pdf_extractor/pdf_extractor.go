package pdf_extractor

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func PdfToImages(tmpDir string, pdfPath string, outPrefix string) ([]string, error) {
	baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))

	pdfDir := filepath.Join(tmpDir, baseName)
	err := os.Mkdir(pdfDir, 0755)
	if err != nil {
		return nil, err
	}

	fullOutPrefix := filepath.Join(pdfDir, outPrefix)

	cmd := exec.Command("pdftoppm", "-png", pdfPath, fullOutPrefix, "-scale-to", "512")
	err = cmd.Run()
	if err != nil {
		//os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to run pdftoppm: %w", err)
	}

	var images []string
	err = filepath.WalkDir(pdfDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".png") {
			images = append(images, path)
		}
		return nil
	})
	if err != nil {
		//os.RemoveAll(tmpDir)
		return nil, err
	}

	sort.Strings(images)
	return images, nil
}
