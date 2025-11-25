package doc_extractor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/previewer/pdf_extractor"
)

func DocumentToImages(tmpDir string, docxPath string, outPrefix string) ([]string, error) {
	baseName := strings.TrimSuffix(filepath.Base(docxPath), filepath.Ext(docxPath))

	docxDir := filepath.Join(tmpDir, baseName)
	err := os.Mkdir(docxDir, 0755)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("unoconv", "-o", docxDir, "-f", "pdf", docxPath)
	err = cmd.Run()
	if err != nil {
		//os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to run unoconv: %w", err)
	}

	return pdf_extractor.PdfToImages(docxDir, fmt.Sprintf("%s.%s", docxDir, "pdf"), outPrefix)
}
