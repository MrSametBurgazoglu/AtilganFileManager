package create_popup

import (
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type CreatePopover struct {
	*gtk.Popover
	NewFileButton      *gtk.Button
	NewDirectoryButton *gtk.Button
	CurrentPath        string
}

func NewCreatePopover(mainWindow *gtk.Window, pathChanged func(string)) *CreatePopover {
	cp := new(CreatePopover)

	popover := gtk.NewPopover()
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	popover.SetChild(box)

	newFileButton := gtk.NewButtonFromIconName("document-new-symbolic")
	newFileButton.ConnectClicked(func() {
		fileSelector := NewFileSelector(cp.CurrentPath, pathChanged)
		fileSelector.SetVisible(true)
		fileSelector.SetTransientFor(mainWindow)
		fileSelector.SetModal(true)
	})
	box.Append(newFileButton)

	newDirectoryButton := gtk.NewButtonFromIconName("folder-new-symbolic")
	newDirectoryButton.ConnectClicked(func() {
		dirSelector := NewDirectorySelector(cp.CurrentPath, pathChanged)
		dirSelector.SetVisible(true)
		dirSelector.SetTransientFor(mainWindow)
		dirSelector.SetModal(true)
	})
	box.Append(newDirectoryButton)
	cp.Popover = popover
	cp.NewFileButton = newFileButton
	cp.NewDirectoryButton = newDirectoryButton

	return cp
}
