package rename_popup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type BulkRenameWindow struct {
	*gtk.Window
	Paths        []string
	Dir          string
	PathChanged  func(string)
	Entries      []*gtk.Entry
	PreviewLabel *gtk.Label
}

func NewBulkRenameWindow(dir string, paths []string, pathChanged func(string)) *BulkRenameWindow {
	win := &BulkRenameWindow{
		Window:      gtk.NewWindow(),
		Paths:       paths,
		Dir:         dir,
		PathChanged: pathChanged,
	}

	win.SetTitle("Bulk Rename")
	win.SetDefaultSize(600, 400)

	mainBox := gtk.NewBox(gtk.OrientationVertical, 12)
	mainBox.SetMarginTop(12)
	mainBox.SetMarginBottom(12)
	mainBox.SetMarginStart(12)
	mainBox.SetMarginEnd(12)
	win.SetChild(mainBox)

	listScroll := gtk.NewScrolledWindow()
	listScroll.SetVExpand(true)
	mainBox.Append(listScroll)

	listBox := gtk.NewBox(gtk.OrientationVertical, 6)
	listScroll.SetChild(listBox)

	for _, p := range paths {
		row := gtk.NewBox(gtk.OrientationHorizontal, 12)
		oldName := filepath.Base(p)
		
		oldLabel := gtk.NewLabel(oldName)
		oldLabel.SetHExpand(true)
		oldLabel.SetHAlign(gtk.AlignStart)
		row.Append(oldLabel)

		arrow := gtk.NewImageFromIconName("go-next-symbolic")
		row.Append(arrow)

		entry := gtk.NewEntry()
		entry.SetText(oldName)
		entry.SetHExpand(true)
		row.Append(entry)
		win.Entries = append(win.Entries, entry)

		listBox.Append(row)
	}

	btnBox := gtk.NewBox(gtk.OrientationHorizontal, 12)
	btnBox.SetHAlign(gtk.AlignEnd)
	mainBox.Append(btnBox)

	cancelBtn := gtk.NewButtonWithLabel("Cancel")
	cancelBtn.ConnectClicked(func() {
		win.Close()
	})
	btnBox.Append(cancelBtn)

	renameBtn := gtk.NewButtonWithLabel("Rename All")
	renameBtn.AddCSSClass("suggested-action")
	renameBtn.ConnectClicked(win.executeRename)
	btnBox.Append(renameBtn)

	return win
}

func (w *BulkRenameWindow) executeRename() {
	for i, p := range w.Paths {
		newName := w.Entries[i].Text()
		if newName == "" || newName == filepath.Base(p) {
			continue
		}

		newPath := filepath.Join(w.Dir, newName)
		err := os.Rename(p, newPath)
		if err != nil {
			fmt.Printf("Error renaming %s to %s: %v\n", p, newPath, err)
		}
	}

	if w.PathChanged != nil {
		w.PathChanged("")
	}
	w.Close()
}
