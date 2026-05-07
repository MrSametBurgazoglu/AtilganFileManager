package previewer

import (
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type FilePreviewer struct {
	*gtk.Box
	Path             string
	NameLabel        *gtk.Label
	TypeLabel        *gtk.Label
	SizeLabel        *gtk.Label
	ModifiedLabel    *gtk.Label
	PermissionsLabel *gtk.Label
	OwnerLabel       *gtk.Label
	ThumbnailImage   *gtk.Image
}

func NewFilePreviewer() *FilePreviewer {
	fp := new(FilePreviewer)

	box := gtk.NewBox(gtk.OrientationVertical, 16)
	box.SetVExpand(true)
	box.SetHExpand(true)
	box.SetHAlign(gtk.AlignFill)
	box.SetVAlign(gtk.AlignStart)
	box.SetMarginTop(24)
	box.SetMarginBottom(24)
	box.SetMarginStart(16)
	box.SetMarginEnd(16)

	thumbnailImage := gtk.NewImage()
	thumbnailImage.SetPixelSize(96)
	thumbnailImage.SetMarginBottom(8)

	nameTitle := gtk.NewLabel("")
	nameTitle.SetWrap(true)
	nameTitle.SetJustify(gtk.JustifyCenter)
	nameTitle.AddCSSClass("preview-title")
	nameTitle.SetHAlign(gtk.AlignCenter)

	infoGrid := gtk.NewGrid()
	infoGrid.AddCSSClass("preview-info-box")
	infoGrid.SetHAlign(gtk.AlignFill)
	infoGrid.SetColumnSpacing(12)
	infoGrid.SetRowSpacing(8)

	rowCount := 0
	createRow := func(title string, valueWidget gtk.Widgetter) {
		row := gtk.NewBox(gtk.OrientationHorizontal, 12)
		titleLabel := gtk.NewLabel(title)
		titleLabel.AddCSSClass("preview-label-title")
		titleLabel.SetHAlign(gtk.AlignStart)
		
		spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
		spacer.SetHExpand(true)
		
		row.Append(titleLabel)
		row.Append(spacer)
		row.Append(valueWidget)
		
		infoGrid.Attach(row, 0, rowCount, 1, 1)
		rowCount++
	}

	typeLabel := gtk.NewLabel("")
	typeLabel.AddCSSClass("preview-label-value")

	sizeLabel := gtk.NewLabel("")
	sizeLabel.AddCSSClass("preview-label-value")

	modifiedLabel := gtk.NewLabel("")
	modifiedLabel.AddCSSClass("preview-label-value")

	permissionsLabel := gtk.NewLabel("")
	permissionsLabel.AddCSSClass("preview-label-value")

	ownerLabel := gtk.NewLabel("")
	ownerLabel.AddCSSClass("preview-label-value")

	createRow("Type:", typeLabel)
	createRow("Size:", sizeLabel)
	createRow("Modified:", modifiedLabel)
	createRow("Permissions:", permissionsLabel)
	createRow("Owner:", ownerLabel)

	box.Append(thumbnailImage)
	box.Append(nameTitle)
	box.Append(infoGrid)

	fp.Box = box
	fp.NameLabel = nameTitle
	fp.TypeLabel = typeLabel
	fp.SizeLabel = sizeLabel
	fp.ModifiedLabel = modifiedLabel
	fp.PermissionsLabel = permissionsLabel
	fp.OwnerLabel = ownerLabel
	fp.ThumbnailImage = thumbnailImage

	return fp
}

func (fp *FilePreviewer) SetFile(filePath string, fileInfo os.FileInfo) {
	fp.Path = filePath
	fp.NameLabel.SetText(fileInfo.Name())
	fp.TypeLabel.SetText(fileops.GetFileDescription(fileInfo.Name()))
	fp.SizeLabel.SetText(fileops.GetFileSizeAsString(fileInfo.Size()))
	fp.ModifiedLabel.SetText(fileops.GetModifiedTimeAsString(fileInfo.ModTime()))
	fp.PermissionsLabel.SetText(fileInfo.Mode().String())

	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		u, err := user.LookupId(strconv.Itoa(int(stat.Uid)))
		if err == nil {
			fp.OwnerLabel.SetText(u.Username)
		} else {
			fp.OwnerLabel.SetText(strconv.Itoa(int(stat.Uid)))
		}
	} else {
		fp.OwnerLabel.SetText("Unknown")
	}

	texture, err := thumbnail.Generate(filePath)
	if err == nil {
		fp.ThumbnailImage.SetFromPaintable(texture)
	} else {
		fp.ThumbnailImage.SetFromIconName("text-x-generic-symbolic")
	}
}
