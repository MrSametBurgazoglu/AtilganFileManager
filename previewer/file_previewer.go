package previewer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const maxPathLength = 48

type FilePreviewer struct {
	*gtk.Box
	Path           string
	NameLabel      *gtk.Label
	SizeLabel      *gtk.Label
	ModifiedLabel  *gtk.Label
	PathLabel      *gtk.Label
	ThumbnailImage *gtk.Image
}

func NewFilePreviewer() *FilePreviewer {
	fp := new(FilePreviewer)

	box := gtk.NewBox(gtk.OrientationVertical, 20)
	box.SetVExpand(true)
	box.SetHExpand(true)
	box.SetHAlign(gtk.AlignCenter)
	box.SetVAlign(gtk.AlignCenter)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)

	thumbnailImage := gtk.NewImage()
	thumbnailImage.SetPixelSize(256)

	infoBox := gtk.NewBox(gtk.OrientationVertical, 0)
	infoBox.AddCSSClass("preview-info-box")
	infoBox.SetHAlign(gtk.AlignCenter)
	infoBox.SetVAlign(gtk.AlignCenter)
	infoBox.SetMarginTop(20)
	infoBox.SetMarginBottom(20)
	infoBox.SetMarginStart(20)
	infoBox.SetMarginEnd(20)

	nameBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	nameBox.AddCSSClass("preview-info-item")
	namePlaceHolder := gtk.NewLabel("Name:")
	namePlaceHolder.AddCSSClass("preview-label")
	nameBox.Append(namePlaceHolder)
	nameLabel := gtk.NewLabel("")
	nameLabel.AddCSSClass("preview-label")
	nameLabel.SetHAlign(gtk.AlignBaselineCenter)
	nameBox.Append(nameLabel)
	infoBox.Append(nameBox)

	pathBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	pathBox.AddCSSClass("preview-info-item")
	pathPlaceHolder := gtk.NewLabel("Path:")
	pathPlaceHolder.AddCSSClass("preview-label")
	pathBox.Append(pathPlaceHolder)
	pathLabel := gtk.NewLabel("")
	pathLabel.AddCSSClass("preview-label")
	pathLabel.SetHAlign(gtk.AlignBaselineCenter)
	pathBox.Append(pathLabel)
	pathCopyButton := gtk.NewButtonFromIconName("edit-copy-symbolic")
	pathCopyButton.Connect("clicked", func() {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		clipboard.SetText(fp.Path)
	})
	pathBox.Append(pathCopyButton)
	infoBox.Append(pathBox)

	sizeBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	sizeBox.AddCSSClass("preview-info-item")
	sizePlaceHolder := gtk.NewLabel("Size:")
	sizePlaceHolder.AddCSSClass("preview-label")
	sizeBox.Append(sizePlaceHolder)
	sizeLabel := gtk.NewLabel("")
	sizeLabel.AddCSSClass("preview-label")
	sizeLabel.SetHAlign(gtk.AlignBaselineCenter)
	sizeBox.Append(sizeLabel)
	infoBox.Append(sizeBox)

	modifiedBox := gtk.NewBox(gtk.OrientationHorizontal, 5)
	modifiedBox.AddCSSClass("preview-info-item")
	modifiedPlaceHolder := gtk.NewLabel("Modified:")
	modifiedPlaceHolder.AddCSSClass("preview-label")
	modifiedBox.Append(modifiedPlaceHolder)
	modifiedLabel := gtk.NewLabel("")
	modifiedLabel.AddCSSClass("preview-label")
	modifiedLabel.SetHAlign(gtk.AlignBaselineCenter)
	modifiedBox.Append(modifiedLabel)
	infoBox.Append(modifiedBox)

	box.Append(thumbnailImage)
	box.Append(infoBox)

	fp.Box = box
	fp.NameLabel = nameLabel
	fp.SizeLabel = sizeLabel
	fp.PathLabel = pathLabel
	fp.ModifiedLabel = modifiedLabel
	fp.ThumbnailImage = thumbnailImage

	return fp
}

func (fp *FilePreviewer) SetFile(filePath string, fileInfo os.FileInfo) {
	fp.Path = filePath
	fp.NameLabel.SetText(fileInfo.Name())
	path := filePath
	pathLength := len(path)
	if pathLength > maxPathLength {
		currentPath := ".../"
		dirs := strings.Split(path, "/")
		startIndex := 0
		for startIndex < len(dirs) {
			if pathLength-len(dirs[startIndex]) < maxPathLength {
				break
			}
			pathLength -= len(dirs[startIndex])
			startIndex++
		}
		pathLeft := filepath.Join(dirs[startIndex:]...)
		path = currentPath + pathLeft
	}
	fp.PathLabel.SetText(path)
	fp.SizeLabel.SetText(fileops.GetFileSizeAsString(fileInfo.Size()))
	fp.ModifiedLabel.SetText(fileops.GetModifiedTimeAsString(fileInfo.ModTime()))
	texture, err := thumbnail.Generate(filePath)
	if err == nil {
		fp.ThumbnailImage.SetFromPaintable(texture)
	} else {
		fp.ThumbnailImage.SetFromIconName("text-x-generic-symbolic")
	}
}
