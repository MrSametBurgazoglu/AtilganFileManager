package previewer

import (
	"fmt"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/MrSametBurgazoglu/atilgan/trash"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type TrashPreviewer struct {
	*gtk.Box
	CurrentItem           string
	ThumbnailImage        *gtk.Image
	NameLabel             *gtk.Label
	OriginalLocationLabel *gtk.Label
	DeleteTimeLabel       *gtk.Label
	RestoreButton         *gtk.Button
}

func NewTrashPreviewer(pathUpdate func()) *TrashPreviewer {
	tp := new(TrashPreviewer)

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
	thumbnailImage.SetPixelSize(144)
	thumbnailImage.SetMarginBottom(8)

	nameTitle := gtk.NewLabel("")
	nameTitle.SetWrap(true)
	nameTitle.SetJustify(gtk.JustifyCenter)
	nameTitle.AddCSSClass("preview-title")
	nameTitle.SetHAlign(gtk.AlignCenter)

	infoBox := gtk.NewBox(gtk.OrientationVertical, 12)
	infoBox.AddCSSClass("preview-info-box")
	infoBox.SetHAlign(gtk.AlignFill)

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
		infoBox.Append(row)
	}

	originalLocationLabel := gtk.NewLabel("")
	originalLocationLabel.AddCSSClass("preview-label-value")
	originalLocationLabel.SetMaxWidthChars(30)
	originalLocationLabel.SetEllipsize(pango.EllipsizeEnd)

	deleteTimeLabel := gtk.NewLabel("")
	deleteTimeLabel.AddCSSClass("preview-label-value")

	createRow("Original Path:", originalLocationLabel)
	createRow("Deleted At:", deleteTimeLabel)

	restoreButton := gtk.NewButtonWithLabel("Restore")
	restoreButton.SetMarginTop(12)
	restoreButton.AddCSSClass("suggested-action")
	restoreButton.ConnectClicked(func() {
		if tp.CurrentItem != "" {
			trash.Restore(tp.CurrentItem)
			pathUpdate()
		}
	})

	box.Append(thumbnailImage)
	box.Append(nameTitle)
	box.Append(infoBox)
	box.Append(restoreButton)

	tp.Box = box
	tp.ThumbnailImage = thumbnailImage
	tp.NameLabel = nameTitle
	tp.OriginalLocationLabel = originalLocationLabel
	tp.DeleteTimeLabel = deleteTimeLabel
	tp.RestoreButton = restoreButton

	return tp
}

func (tp *TrashPreviewer) SetFilePath(filePath string) {
	fileName := filepath.Base(filePath)
	tp.CurrentItem = fileName
	info, err := trash.GetItemInfo(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	tp.NameLabel.SetText(info.Name)
	tp.OriginalLocationLabel.SetText(info.OriginalPath)
	tp.OriginalLocationLabel.SetTooltipText(info.OriginalPath)
	tp.DeleteTimeLabel.SetText(info.DeletionDate)
	tp.RestoreButton.SetLabel(fmt.Sprintf("Restore %s", info.Name))

	realPath, err := trash.GetTrashFilePath(fileName)
	if err == nil {
		texture, err := thumbnail.Generate(realPath)
		if err == nil {
			tp.ThumbnailImage.SetFromPaintable(texture)
		} else {
			tp.ThumbnailImage.SetFromIconName("text-x-generic-symbolic")
		}
	} else {
		tp.ThumbnailImage.SetFromIconName("text-x-generic-symbolic")
	}
}
