package previewer_panel

import (
	"os"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/previewer"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PreviewPanel struct {
	*gtk.Stack
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
	pp.SetHExpand(true)

	emptyPreviewer := gtk.NewLabel("Empty Directory")

	pp.AddTitled(emptyPreviewer, "emptypreviewer", "Empty Previewer")
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
			pp.dirPreviewer.SetPath(filePath)
			pp.SetVisibleChildName("dirviewer")
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

	if isImage(info.Name()) {
		pp.imagePreviewer.SetImage(pp.filePath, info)
		pp.SetVisibleChildName("imagepreviewer")
	} else if isText(info.Name()) {
		pp.textPreviewer.SetText(pp.filePath, info)
		pp.SetVisibleChildName("textpreviewer")
	} else if isCode(info.Name()) {
		pp.codePreviewer.SetText(pp.filePath, info)
		pp.SetVisibleChildName("codepreviewer")
	} else if isMedia(info.Name()) {
		pp.mediaPreviewer.SetMedia(pp.filePath, info)
		pp.SetVisibleChildName("mediapreviewer")
	} else if isDocument(info.Name()) {
		pp.documentPreviewer.SetDocument(pp.filePath, info)
		pp.SetVisibleChildName("documentpreviewer")
	} else {
		pp.filePreviewer.SetFile(pp.filePath, info)
		pp.SetVisibleChildName("filepreviewer")
	}
}

func isImage(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".png") ||
		strings.HasSuffix(fileName, ".jpg") ||
		strings.HasSuffix(fileName, ".jpeg") ||
		strings.HasSuffix(fileName, ".gif") ||
		strings.HasSuffix(fileName, ".webp") ||
		strings.HasSuffix(fileName, ".svg")
}

func isText(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".txt") ||
		strings.HasSuffix(fileName, ".mod") ||
		strings.HasSuffix(fileName, ".sum")
}

func isMedia(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".mp3") || strings.HasSuffix(fileName, ".mp4")
}

func isDocument(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".pdf") ||
		strings.HasSuffix(fileName, ".epub") ||
		strings.HasSuffix(fileName, ".mobi") ||
		strings.HasSuffix(fileName, ".docx") ||
		strings.HasSuffix(fileName, ".xlsx") ||
		strings.HasSuffix(fileName, ".pptx")
}

func isCode(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".go") ||
		strings.HasSuffix(fileName, ".json") ||
		strings.HasSuffix(fileName, ".yaml") ||
		strings.HasSuffix(fileName, ".yml") ||
		strings.HasSuffix(fileName, ".env") ||
		strings.HasSuffix(fileName, "dockerfile") ||
		strings.HasSuffix(fileName, ".js") ||
		strings.HasSuffix(fileName, ".ts") ||
		strings.HasSuffix(fileName, ".py") ||
		strings.HasSuffix(fileName, ".java") ||
		strings.HasSuffix(fileName, ".c") ||
		strings.HasSuffix(fileName, ".cpp") ||
		strings.HasSuffix(fileName, ".h") ||
		strings.HasSuffix(fileName, ".hpp") ||
		strings.HasSuffix(fileName, ".rs") ||
		strings.HasSuffix(fileName, ".rb") ||
		strings.HasSuffix(fileName, ".php") ||
		strings.HasSuffix(fileName, ".swift") ||
		strings.HasSuffix(fileName, ".kt") ||
		strings.HasSuffix(fileName, ".kts") ||
		strings.HasSuffix(fileName, ".sh") ||
		strings.HasSuffix(fileName, ".bat") ||
		strings.HasSuffix(fileName, ".md")
}
