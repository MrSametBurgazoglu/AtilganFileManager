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
	box := gtk.NewBox(gtk.OrientationVertical, 4)
	box.SetMarginTop(4)
	box.SetMarginBottom(4)
	box.SetMarginStart(4)
	box.SetMarginEnd(4)
	popover.SetChild(box)

	createBox := gtk.NewBox(gtk.OrientationHorizontal, 4)
	createBox.SetHomogeneous(true)
	box.Append(createBox)

	createButton := func(iconName, labelText string, action func()) *gtk.Button {
		btn := gtk.NewButton()
		btn.AddCSSClass("flat")
		
		btnBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
		btnBox.SetMarginStart(8)
		btnBox.SetMarginEnd(16)
		btnBox.SetMarginTop(4)
		btnBox.SetMarginBottom(4)
		
		icon := gtk.NewImageFromIconName(iconName)
		label := gtk.NewLabel(labelText)
		label.SetHAlign(gtk.AlignStart)
		
		btnBox.Append(icon)
		btnBox.Append(label)
		btn.SetChild(btnBox)
		
		btn.ConnectClicked(func() {
			action()
			popover.Popdown()
		})
		return btn
	}

	newFileButton := createButton("document-new-symbolic", "New File", func() {
		fileSelector := NewFileSelector(cp.CurrentPath, pathChanged)
		fileSelector.SetVisible(true)
		fileSelector.SetTransientFor(mainWindow)
		fileSelector.SetModal(true)
	})
	createBox.Append(newFileButton)

	newDirectoryButton := createButton("folder-new-symbolic", "New Folder", func() {
		dirSelector := NewDirectorySelector(cp.CurrentPath, pathChanged)
		dirSelector.SetVisible(true)
		dirSelector.SetTransientFor(mainWindow)
		dirSelector.SetModal(true)
	})
	createBox.Append(newDirectoryButton)
	
	cp.Popover = popover
	cp.NewFileButton = newFileButton
	cp.NewDirectoryButton = newDirectoryButton

	return cp
}
