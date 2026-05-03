package previewer_panel

import (
	"os"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PreviewPanel struct {
	*gtk.Stack
	imageDirPreviewer  *previewer.ImageDirPreviewer
	dirPreviewer       *previewer.DirPreviewer
	filePreviewer      *previewer.FilePreviewer
	imagePreviewer     *previewer.ImagePreviewer
	textPreviewer      *previewer.TextPreviewer
	codePreviewer      *previewer.CodePreviewer
	mediaPreviewer     *previewer.MediaPreviewer
	documentPreviewer  *previewer.DocumentPreviewer
	trashPreviewer     *previewer.TrashPreviewer
	filePath           string
	specialPathManager *special_path.SpecialPathManager
}

func NewPreviewPanel(path string, changePath func(string), specialPathManager *special_path.SpecialPathManager) *PreviewPanel {
	pp := &PreviewPanel{
		Stack:              gtk.NewStack(),
		imageDirPreviewer:  previewer.NewImageDirPreviewer(path, changePath, specialPathManager),
		dirPreviewer:       previewer.NewDirPreviewer(path, changePath, specialPathManager),
		filePreviewer:      previewer.NewFilePreviewer(),
		imagePreviewer:     previewer.NewImagePreviewer(),
		textPreviewer:      previewer.NewTextPreviewer(),
		codePreviewer:      previewer.NewCodePreviewer(),
		mediaPreviewer:     previewer.NewMediaPreviewer(),
		documentPreviewer:  previewer.NewDocumentPreviewer(),
		trashPreviewer:     previewer.NewTrashPreviewer(func() { changePath("trash://") }),
		specialPathManager: specialPathManager,
	}
	pp.AddCSSClass("preview-panel")

	emptyPreviewer := gtk.NewLabel("Empty Directory")
	emptyPreviewer.AddCSSClass("preview-title")
	emptyPreviewer.SetHAlign(gtk.AlignCenter)
	emptyPreviewer.SetVAlign(gtk.AlignCenter)

	pp.AddTitled(emptyPreviewer, "emptypreviewer", "Empty Previewer")
	pp.AddTitled(pp.imageDirPreviewer, "imagedirviewer", "Image Directory Viewer")
	pp.AddTitled(pp.dirPreviewer, "dirviewer", "Directory Viewer")
	pp.AddTitled(pp.filePreviewer, "filepreviewer", "File Previewer")
	pp.AddTitled(pp.imagePreviewer, "imagepreviewer", "Image Previewer")
	pp.AddTitled(pp.textPreviewer, "textpreviewer", "Text Previewer")
	pp.AddTitled(pp.codePreviewer, "codepreviewer", "Code Previewer")
	pp.AddTitled(pp.mediaPreviewer, "mediapreviewer", "Media Previewer")
	pp.AddTitled(pp.documentPreviewer, "documentpreviewer", "Document Previewer")
	pp.AddTitled(pp.trashPreviewer, "trashpreviewer", "Trash Previewer")

	pp.SetVExpand(true)
	return pp
}

func (pp *PreviewPanel) Update(filePath string) {
	pp.mediaPreviewer.Close()
	pp.documentPreviewer.Close()
	pp.filePath = filePath

	if filePath == "" {
		pp.SetVisibleChildName("emptypreviewer")
		return
	}

	if strings.HasPrefix(filePath, "trash://") {
		pp.SetVisibleChildName("trashpreviewer")
		pp.trashPreviewer.SetFilePath(filePath)
		return
	}
	if strings.HasPrefix(filePath, "tags://") {
		pp.dirPreviewer.SetPath(filePath)
		pp.SetVisibleChildName("dirviewer")
		return
	}

	info, err := os.Stat(filePath)
	if err == nil {
		if info.IsDir() {
			// Smart switching logic
			entries, err := os.ReadDir(filePath)
			if err == nil {
				imageCount := 0
				totalCount := 0
				for _, entry := range entries {
					if !entry.IsDir() {
						totalCount++
						if fileops.IsImage(entry.Name()) {
							imageCount++
						}
					}
				}
				if totalCount > 0 && imageCount > totalCount/2 {
					pp.imageDirPreviewer.SetPath(filePath)
					pp.SetVisibleChildName("imagedirviewer")
				} else {
					pp.dirPreviewer.SetPath(filePath)
					pp.SetVisibleChildName("dirviewer")
				}
			} else {
				pp.dirPreviewer.SetPath(filePath)
				pp.SetVisibleChildName("dirviewer")
			}
		} else {
			pp.filePreviewer.SetFile(filePath, info)
			pp.SetVisibleChildName("filepreviewer")
			pp.specialPathManager.AddRecentPath(filePath)
		}
	}
}

func (pp *PreviewPanel) ShowSpecificPreviewer() {
	if pp.filePath == "" {
		return
	}
	info, err := os.Stat(pp.filePath)
	if err != nil {
		return
	}
	if info.IsDir() {
		return
	}

	pp.specialPathManager.AddRecentPath(pp.filePath)

	if fileops.IsImage(info.Name()) {
		pp.imagePreviewer.SetImage(pp.filePath, info)
		pp.SetVisibleChildName("imagepreviewer")
	} else if fileops.IsText(info.Name()) {
		pp.textPreviewer.SetText(pp.filePath, info)
		pp.SetVisibleChildName("textpreviewer")
	} else if fileops.IsCode(info.Name()) {
		pp.codePreviewer.SetText(pp.filePath, info)
		pp.SetVisibleChildName("codepreviewer")
	} else if fileops.IsMedia(info.Name()) {
		pp.mediaPreviewer.SetMedia(pp.filePath, info)
		pp.SetVisibleChildName("mediapreviewer")
	} else if fileops.IsDocument(info.Name()) {
		pp.documentPreviewer.SetDocument(pp.filePath, info)
		pp.SetVisibleChildName("documentpreviewer")
	} else {
		pp.filePreviewer.SetFile(pp.filePath, info)
		pp.SetVisibleChildName("filepreviewer")
	}
}
