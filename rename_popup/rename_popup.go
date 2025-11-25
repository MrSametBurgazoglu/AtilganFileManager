package rename_popup

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type RenameWindow struct {
	*gtk.Window
	Entry        *gtk.Entry
	BasePath     string
	SelectedPath string
}

func NewRenameWindow(basePath string, selectedPath string) *RenameWindow {
	rw := &RenameWindow{
		Window:       gtk.NewWindow(),
		Entry:        gtk.NewEntry(),
		BasePath:     basePath,
		SelectedPath: selectedPath,
	}

	rw.SetTitle("Rename")
	rw.SetDefaultSize(400, 50)
	rw.SetModal(true)

	box := gtk.NewBox(gtk.OrientationVertical, 5)
	rw.SetChild(box)

	rw.Entry.SetText(filepath.Base(selectedPath))
	box.Append(rw.Entry)

	rw.Entry.Connect("activate", func() {
		if rw.isEntryValid() {
			newPath := filepath.Join(rw.BasePath, rw.GetNewName())
			err := os.Rename(rw.SelectedPath, newPath)
			if err != nil {
				log.Printf("Error renaming %s to %s: %v", rw.SelectedPath, newPath, err)
			} else {
				rw.Destroy()
			}
		} else {
			rw.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
		}
	})

	return rw
}

func (rw *RenameWindow) GetNewName() string {
	return rw.Entry.Text()
}

func (rw *RenameWindow) isEntryValid() bool {
	searchText := rw.Entry.Text()

	hasIllegalChars := false
	if strings.ContainsAny(searchText, "/\\:*?\"<>|") || len(searchText) == 0 {
		hasIllegalChars = true
	}

	newPath := filepath.Join(rw.BasePath, searchText)
	_, err := os.Stat(newPath)
	alreadyExists := !os.IsNotExist(err)

	if searchText == filepath.Base(rw.SelectedPath) {
		alreadyExists = false
	}

	return !hasIllegalChars && !alreadyExists
}
