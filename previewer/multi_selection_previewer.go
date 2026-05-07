package previewer

import (
	"fmt"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type MultiSelectionPreviewer struct {
	*gtk.Box
	TitleLabel *gtk.Label
	PathsBox   *gtk.Box
	CountLabel *gtk.Label
	Icon       *gtk.Image
}

func NewMultiSelectionPreviewer() *MultiSelectionPreviewer {
	mp := &MultiSelectionPreviewer{
		Box: gtk.NewBox(gtk.OrientationVertical, 16),
	}

	mp.SetVExpand(true)
	mp.SetHExpand(true)
	mp.SetHAlign(gtk.AlignFill)
	mp.SetVAlign(gtk.AlignFill)
	mp.SetMarginTop(24)
	mp.SetMarginBottom(24)
	mp.SetMarginStart(16)
	mp.SetMarginEnd(16)

	mp.Icon = gtk.NewImage()
	mp.Icon.SetPixelSize(32)
	mp.Icon.SetMarginBottom(8)
	mp.Icon.SetFromIconName("emblem-symbolic-link") // Or something else

	mp.TitleLabel = gtk.NewLabel("Multiple Selection")
	mp.TitleLabel.AddCSSClass("preview-title")
	mp.TitleLabel.SetHAlign(gtk.AlignCenter)

	mp.CountLabel = gtk.NewLabel("")
	mp.CountLabel.AddCSSClass("preview-label-title")
	mp.CountLabel.SetMarginBottom(12)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)
	scrolled.SetMinContentHeight(100)
	scrolled.AddCSSClass("view")

	mp.PathsBox = gtk.NewBox(gtk.OrientationVertical, 4)
	mp.PathsBox.SetMarginTop(8)
	mp.PathsBox.SetMarginBottom(8)
	mp.PathsBox.SetMarginStart(8)
	mp.PathsBox.SetMarginEnd(8)
	scrolled.SetChild(mp.PathsBox)

	mp.Append(mp.Icon)
	mp.Append(mp.TitleLabel)
	mp.Append(mp.CountLabel)
	mp.Append(scrolled)

	return mp
}

func (mp *MultiSelectionPreviewer) SetFiles(paths []string) {
	mp.CountLabel.SetText(fmt.Sprintf("%d items selected", len(paths)))
	
	for child := mp.PathsBox.FirstChild(); child != nil; child = mp.PathsBox.FirstChild() {
		mp.PathsBox.Remove(child)
	}

	for _, path := range paths {
		name := filepath.Base(path)
		label := gtk.NewLabel(name)
		label.SetHAlign(gtk.AlignStart)
		label.SetEllipsize(1) // pango.EllipsizeEnd
		label.SetTooltipText(path)
		mp.PathsBox.Append(label)
	}
}
