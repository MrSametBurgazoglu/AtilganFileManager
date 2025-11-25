package tag_popup

import (
	"github.com/MrSametBurgazoglu/atilgan/tag"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type TagPopup struct {
	*gtk.Window
	entry      *gtk.Entry
	tagManager *tag.TagManager
	path       string
}

func NewTagPopup(parent *gtk.Window, tagManager *tag.TagManager, path string) *TagPopup {
	popup := &TagPopup{
		Window:     gtk.NewWindow(),
		entry:      gtk.NewEntry(),
		tagManager: tagManager,
		path:       path,
	}

	popup.SetTransientFor(parent)
	popup.SetModal(true)
	popup.SetTitle("Add Tag")

	box := gtk.NewBox(gtk.OrientationVertical, 6)
	box.SetMarginTop(12)
	box.SetMarginBottom(12)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)
	popup.SetChild(box)

	box.Append(popup.entry)

	addButton := gtk.NewButtonWithLabel("Add")
	addButton.ConnectClicked(func() {
		tag := popup.entry.Text()
		if tag != "" {
			popup.tagManager.AddTag(popup.path, tag)
			popup.Close()
		}
	})
	box.Append(addButton)

	popup.entry.ConnectActivate(func() {
		addButton.Activate()
	})

	return popup
}
