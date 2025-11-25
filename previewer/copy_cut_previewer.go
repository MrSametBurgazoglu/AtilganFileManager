package previewer

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CopyCutPreviewer struct {
	*gtk.Box
	OperationLabel *gtk.Label
	IsCut          bool
	Paths          *gtk.Box
	SizeLabel      *gtk.Label
}

func NewCopyCutPreviewer() *CopyCutPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetHExpand(true)
	nameLabel := gtk.NewLabel("")
	operationLabel := gtk.NewLabel("")
	paths := gtk.NewBox(gtk.OrientationVertical, 0)
	sizeLabel := gtk.NewLabel("")

	box.Append(nameLabel)
	box.Append(operationLabel)
	box.Append(paths)
	box.Append(sizeLabel)

	return &CopyCutPreviewer{
		Box:            box,
		OperationLabel: operationLabel,
		Paths:          paths,
		SizeLabel:      sizeLabel,
	}
}

func (cp *CopyCutPreviewer) SetFiles(paths []string) {
	if cp.IsCut {
		cp.OperationLabel.SetText("Cutting File/Directory")
	} else {
		cp.OperationLabel.SetText("Copying File/Directory")
	}
	for child := cp.Paths.FirstChild(); child != nil; child = cp.Paths.FirstChild() {
		cp.Paths.Remove(child)
	}
	for _, path := range paths {
		pathLabel := gtk.NewLabel(path)
		cp.Paths.Append(pathLabel)
	}
}
