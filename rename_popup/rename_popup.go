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
	rw.SetResizable(false)
	rw.SetDefaultSize(350, -1)
	rw.SetModal(true)

	headerBar := gtk.NewHeaderBar()
	headerBar.SetShowTitleButtons(false)

	titleLabel := gtk.NewLabel("Rename")
	titleLabel.AddCSSClass("title-4")
	headerBar.SetTitleWidget(titleLabel)

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		rw.Destroy()
	})
	headerBar.PackStart(cancelBtn)

	renameBtn := gtk.NewButtonWithLabel("Rename")
	renameBtn.AddCSSClass("suggested-action")
	headerBar.PackEnd(renameBtn)
	rw.SetTitlebar(headerBar)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(16)
	mainBox.SetMarginBottom(16)
	mainBox.SetMarginStart(16)
	mainBox.SetMarginEnd(16)

	entryBox := gtk.NewBox(gtk.OrientationVertical, 4)
	entryLabel := gtk.NewLabel("New Name")
	entryLabel.SetHAlign(gtk.AlignStart)
	entryLabel.AddCSSClass("caption")

	rw.Entry.SetText(filepath.Base(selectedPath))
	rw.Entry.SetHExpand(true)
	
	entryBox.Append(entryLabel)
	entryBox.Append(rw.Entry)
	mainBox.Append(entryBox)

	validate := func() {
		isValid := rw.isEntryValid()
		renameBtn.SetSensitive(isValid)
		if isValid {
			rw.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "")
			rw.Entry.RemoveCSSClass("error")
		} else {
			rw.Entry.SetIconFromIconName(gtk.EntryIconSecondary, "window-close-symbolic")
			rw.Entry.AddCSSClass("error")
		}
	}

	rw.Entry.Connect("notify::text", validate)

	renameAction := func() {
		if rw.isEntryValid() {
			newPath := filepath.Join(rw.BasePath, rw.GetNewName())
			err := os.Rename(rw.SelectedPath, newPath)
			if err != nil {
				log.Printf("Error renaming %s to %s: %v", rw.SelectedPath, newPath, err)
			} else {
				rw.Destroy()
			}
		}
	}

	rw.Entry.Connect("activate", renameAction)
	renameBtn.ConnectClicked(renameAction)

	rw.SetChild(mainBox)
	rw.Entry.GrabFocus()
	validate()

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
