package previewer

import (
	"fmt"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CopyCutPreviewer struct {
	*gtk.Box
	OperationLabel *gtk.Label
	IsCut          bool
	PathsBox       *gtk.Box
	CountLabel     *gtk.Label
	Icon           *gtk.Image
	OnClear        func()
}

func NewCopyCutPreviewer() *CopyCutPreviewer {
	cp := &CopyCutPreviewer{
		Box: gtk.NewBox(gtk.OrientationVertical, 16),
	}

	cp.SetVExpand(true)
	cp.SetHExpand(true)
	cp.SetHAlign(gtk.AlignFill)
	cp.SetVAlign(gtk.AlignStart)
	cp.SetMarginTop(24)
	cp.SetMarginBottom(24)
	cp.SetMarginStart(16)
	cp.SetMarginEnd(16)

	cp.Icon = gtk.NewImage()
	cp.Icon.SetPixelSize(32)
	cp.Icon.SetMarginBottom(8)

	cp.OperationLabel = gtk.NewLabel("")
	cp.OperationLabel.AddCSSClass("preview-title")
	cp.OperationLabel.SetHAlign(gtk.AlignCenter)

	cp.CountLabel = gtk.NewLabel("")
	cp.CountLabel.AddCSSClass("preview-label-title")
	cp.CountLabel.SetMarginBottom(12)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetMinContentHeight(100)
	scrolled.SetMaxContentHeight(300)
	scrolled.AddCSSClass("view")

	cp.PathsBox = gtk.NewBox(gtk.OrientationVertical, 4)
	cp.PathsBox.SetMarginTop(8)
	cp.PathsBox.SetMarginBottom(8)
	cp.PathsBox.SetMarginStart(8)
	cp.PathsBox.SetMarginEnd(8)
	scrolled.SetChild(cp.PathsBox)

	clearButton := gtk.NewButtonWithLabel("Clear Clipboard")
	clearButton.AddCSSClass("suggested-action")
	clearButton.SetMarginTop(16)
	clearButton.ConnectClicked(func() {
		if cp.OnClear != nil {
			cp.OnClear()
		}
	})

	cp.Append(cp.Icon)
	cp.Append(cp.OperationLabel)
	cp.Append(cp.CountLabel)
	cp.Append(scrolled)
	cp.Append(clearButton)

	return cp
}

func (cp *CopyCutPreviewer) SetFiles(paths []string) {
	if cp.IsCut {
		cp.OperationLabel.SetText("Files to Move")
		cp.Icon.SetFromIconName("edit-cut-symbolic")
	} else {
		cp.OperationLabel.SetText("Files to Copy")
		cp.Icon.SetFromIconName("edit-copy-symbolic")
	}

	cp.CountLabel.SetText(fmt.Sprintf("%d items selected", len(paths)))
	
	for child := cp.PathsBox.FirstChild(); child != nil; child = cp.PathsBox.FirstChild() {
		cp.PathsBox.Remove(child)
	}

	for _, path := range paths {
		name := filepath.Base(path)
		label := gtk.NewLabel(name)
		label.SetHAlign(gtk.AlignStart)
		label.SetEllipsize(1) // pango.EllipsizeEnd
		label.SetTooltipText(path)
		cp.PathsBox.Append(label)
	}
}

