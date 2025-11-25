package pathbar

import (
	"path/filepath"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type PathBar struct {
	*gtk.Stack
	PathbarBox      *gtk.Box
	PathBarEntryBox *PathBarEntryBox
	currentPath     string
	previousPath    string

	SetPath func(string)
}

type PathBarEntryBox struct {
	*gtk.Box
	PathEntry *gtk.Entry
}

func NewPathBarEntryBox() *PathBarEntryBox {
	entryBox := new(PathBarEntryBox)
	entryBox.Box = gtk.NewBox(gtk.OrientationHorizontal, 6)
	entryBox.PathEntry = gtk.NewEntry()
	entryBox.Append(entryBox.PathEntry)

	copyButton := gtk.NewButtonFromIconName("edit-copy-symbolic")
	copyButton.ConnectClicked(func() {
		clipboard := gdk.DisplayGetDefault().Clipboard()
		clipboard.SetText(entryBox.PathEntry.Text())
	})
	entryBox.Append(copyButton)

	return entryBox
}

func NewPathBar(setPath func(string)) *PathBar {
	pathBar := &PathBar{
		Stack: gtk.NewStack(),
	}

	pathBar.SetPath = setPath
	pathBar.SetHAlign(gtk.AlignStart)
	pathBar.SetHExpand(true)
	pathBar.SetHAlign(gtk.AlignCenter)

	pathBar.PathbarBox = gtk.NewBox(gtk.OrientationHorizontal, 0)
	pathBar.PathbarBox.AddCSSClass("path-bar")
	pathBar.AddTitled(pathBar.PathbarBox, "pathbar", "Path Bar")

	pathBar.PathBarEntryBox = NewPathBarEntryBox()
	pathBar.AddTitled(pathBar.PathBarEntryBox, "pathentry", "Path Entry")
	pathBar.PathBarEntryBox.PathEntry.ConnectActivate(func() {
		setPath(pathBar.PathBarEntryBox.PathEntry.Text())
		pathBar.SetVisibleChildName("pathbar")
	})

	return pathBar
}

func (pb *PathBar) UpdatePathBar(path string) {
	pb.currentPath = path
	for child := pb.PathbarBox.FirstChild(); child != nil; child = pb.PathbarBox.FirstChild() {
		pb.PathbarBox.Remove(child)
	}

	if strings.Contains(path, "://") {
		name := strings.Split(path, "://")[0]
		name = strings.ToUpper(name[:1]) + name[1:]
		button := gtk.NewToggleButtonWithLabel(name)
		button.SetActive(true)
		button.AddCSSClass("selected")
		button.ConnectToggled(func() {
			pb.PathBarEntryBox.PathEntry.SetText(pb.currentPath)
			pb.SetVisibleChildName("pathentry")
		})
		pb.PathbarBox.Append(button)
		pb.SetVisibleChildName("pathbar")
		return
	}

	usedPath := pb.currentPath
	if pb.previousPath != "" && strings.HasPrefix(pb.previousPath, pb.currentPath) {
		usedPath = pb.previousPath
	} else {
		pb.previousPath = pb.currentPath
	}

	components := strings.Split(usedPath, string(filepath.Separator))
	if len(components) > 0 && components[0] == "" {
		components = components[1:]
	}

	for i, component := range components {
		pathSoFar := "/" + filepath.Join(components[:i+1]...)
		button := gtk.NewToggleButtonWithLabel(component)
		if pathSoFar == pb.currentPath {
			button.SetActive(true)
			button.AddCSSClass("selected")
		}
		button.ConnectToggled(func() {
			if pb.currentPath == "" {
				return
			}
			if pb.currentPath == pathSoFar {
				pb.PathBarEntryBox.PathEntry.SetText(pb.currentPath)
				pb.SetVisibleChildName("pathentry")
				return
			} else if !strings.HasPrefix(pb.previousPath, pathSoFar) {
				pb.previousPath = pb.currentPath
			}
			pb.SetPath(pathSoFar)
		})
		pb.PathbarBox.Append(button)
		if i < len(components)-1 {
			separator := gtk.NewLabel(">")
			separator.AddCSSClass("path-separator")
			pb.PathbarBox.Append(separator)
		}
	}
	pb.SetVisibleChildName("pathbar")
}
