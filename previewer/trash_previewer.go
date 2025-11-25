package previewer

import (
	"fmt"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/trash"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type TrashPreviewer struct {
	*gtk.Box
	NameLabel             *gtk.Label
	OriginalLocationLabel *gtk.Label
	DeleteTimeLabel       *gtk.Label
	RestoreButton         *gtk.Button
}

func NewTrashPreviewer(pathUpdate func()) *TrashPreviewer {
	box := gtk.NewBox(gtk.OrientationVertical, 0)
	box.SetHExpand(true)
	nameLabel := gtk.NewLabel("")
	originalLocationLabel := gtk.NewLabel("")
	deleteTimeLabel := gtk.NewLabel("")
	restoreButton := gtk.NewButtonWithLabel("Restore")
	restoreButton.ConnectClicked(func() {
		trash.Restore(nameLabel.Label())
		pathUpdate()
	})

	box.Append(nameLabel)
	box.Append(originalLocationLabel)
	box.Append(deleteTimeLabel)
	box.Append(restoreButton)

	return &TrashPreviewer{
		Box:                   box,
		NameLabel:             nameLabel,
		OriginalLocationLabel: originalLocationLabel,
		DeleteTimeLabel:       deleteTimeLabel,
		RestoreButton:         restoreButton,
	}
}

func (tp *TrashPreviewer) SetFilePath(filePath string) {
	fileName := filepath.Base(filePath)
	info, err := trash.GetItemInfo(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	tp.NameLabel.SetText(info.Name)
	tp.OriginalLocationLabel.SetText(info.OriginalPath)
	tp.DeleteTimeLabel.SetText(info.DeletionDate)
	tp.RestoreButton.SetLabel(fmt.Sprintf("Restore %s", fileName))
}
