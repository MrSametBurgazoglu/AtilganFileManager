package previewer

import (
	"fmt"
	"os"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/diamondburned/gotk4/pkg/gdkpixbuf/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PictureDetailsPreviewer struct {
	*gtk.Box
	image       *gtk.Picture
	nameLabel   *gtk.Label
	sizeLabel   *gtk.Label
	dimLabel    *gtk.Label
	pathLabel   *gtk.Label
}

func NewPictureDetailsPreviewer() *PictureDetailsPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 10)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	image := gtk.NewPicture()
	image.SetCanShrink(true)
	image.SetHExpand(true)
	image.SetSizeRequest(250, 200)

	infoBox := gtk.NewBox(gtk.OrientationVertical, 4)
	nameLabel := gtk.NewLabel("")
	nameLabel.SetWrap(true)
	nameLabel.AddCSSClass("title-2")
	
	sizeLabel := gtk.NewLabel("")
	dimLabel := gtk.NewLabel("")
	pathLabel := gtk.NewLabel("")
	pathLabel.SetWrap(true)
	pathLabel.AddCSSClass("caption")

	infoBox.Append(nameLabel)
	infoBox.Append(sizeLabel)
	infoBox.Append(dimLabel)
	infoBox.Append(pathLabel)

	box.Append(image)
	box.Append(infoBox)

	return &PictureDetailsPreviewer{
		Box:       box,
		image:     image,
		nameLabel: nameLabel,
		sizeLabel: sizeLabel,
		dimLabel:  dimLabel,
		pathLabel: pathLabel,
	}
}

func (pdp *PictureDetailsPreviewer) SetPicture(filePath string, info os.FileInfo) {
	pdp.image.SetFile(nil) // Clear previous
	pdp.image.SetFilename(filePath)

	pdp.nameLabel.SetLabel(info.Name())
	pdp.sizeLabel.SetLabel(fmt.Sprintf("Size: %s", fileops.GetFileSizeAsString(info.Size())))
	pdp.pathLabel.SetLabel(filePath)

	// Try to get dimensions
	pixbuf, err := gdkpixbuf.NewPixbufFromFile(filePath)
	if err == nil {
		pdp.dimLabel.SetLabel(fmt.Sprintf("Dimensions: %dx%d", pixbuf.Width(), pixbuf.Height()))
	} else {
		pdp.dimLabel.SetLabel("Dimensions: Unknown")
	}
}
