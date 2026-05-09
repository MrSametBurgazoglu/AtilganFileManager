package previewer

import (
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/pango"
)

type RecentPreviewer struct {
	*gtk.Box
	listBox            *gtk.ListBox
	specialPathManager *special_path.SpecialPathManager
	changePath         func(string)
}

func NewRecentPreviewer(specialPathManager *special_path.SpecialPathManager, changePath func(string)) *RecentPreviewer {
	rp := &RecentPreviewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 16),
		specialPathManager: specialPathManager,
		changePath:         changePath,
	}

	rp.SetVExpand(true)
	rp.SetHExpand(true)
	rp.SetMarginTop(24)
	rp.SetMarginBottom(24)
	rp.SetMarginStart(16)
	rp.SetMarginEnd(16)

	title := gtk.NewLabel("Recent Files")
	title.AddCSSClass("preview-title")
	title.SetHAlign(gtk.AlignCenter)

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetVExpand(true)

	rp.listBox = gtk.NewListBox()
	rp.listBox.SetSelectionMode(gtk.SelectionNone)
	rp.listBox.AddCSSClass("boxed-list")
	scrolled.SetChild(rp.listBox)

	rp.Append(title)
	rp.Append(scrolled)

	return rp
}

func (rp *RecentPreviewer) Refresh() {
	// clear list
	for child := rp.listBox.FirstChild(); child != nil; child = rp.listBox.FirstChild() {
		rp.listBox.Remove(child)
	}

	if rp.specialPathManager == nil || rp.specialPathManager.GetRecentManager() == nil {
		return
	}

	recentPaths := rp.specialPathManager.GetRecentManager().GetPaths()
	count := 0
	for _, path := range recentPaths {
		if count >= 10 { // limit to 10 recent items in the previewer panel
			break
		}

		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		count++

		row := gtk.NewListBoxRow()
		box := gtk.NewBox(gtk.OrientationHorizontal, 12)
		box.SetMarginTop(8)
		box.SetMarginBottom(8)
		box.SetMarginStart(12)
		box.SetMarginEnd(12)

		var iconWidget gtk.Widgetter
		texture, err := thumbnail.Generate(path)
		if err == nil {
			pic := gtk.NewPictureForPaintable(texture)
			pic.SetSizeRequest(32, 32)
			pic.SetCanShrink(true)
			// Apply a generic border/radius class if needed, or simply render
			iconWidget = pic
		} else {
			icon := gtk.NewImageFromIconName("text-x-generic-symbolic")
			icon.SetPixelSize(32)
			iconWidget = icon
		}

		vbox := gtk.NewBox(gtk.OrientationVertical, 2)

		nameLabel := gtk.NewLabel(filepath.Base(path))
		nameLabel.SetHAlign(gtk.AlignStart)
		nameLabel.SetEllipsize(pango.EllipsizeEnd)
		nameLabel.AddCSSClass("title-4")

		pathLabel := gtk.NewLabel(filepath.Dir(path))
		pathLabel.SetHAlign(gtk.AlignStart)
		pathLabel.SetEllipsize(pango.EllipsizeStart)
		pathLabel.AddCSSClass("caption")

		vbox.Append(nameLabel)
		vbox.Append(pathLabel)
		vbox.SetHExpand(true)

		box.Append(iconWidget)
		box.Append(vbox)

		// Create a container button to make it clickable
		btn := gtk.NewButton()
		btn.SetChild(box)
		btn.AddCSSClass("flat")
		btn.ConnectClicked(func() {
			rp.changePath(filepath.Dir(path))
		})

		row.SetChild(btn)
		rp.listBox.Append(row)
	}
}
