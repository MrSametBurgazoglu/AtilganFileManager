package previewer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/thumbnail"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type SortOrder int

const (
	SortByName SortOrder = iota
	SortByTime
)

type ImageDirPreviewer struct {
	*gtk.Box
	Path               string
	SortOrder          SortOrder
	changePath         func(string)
	FileViewerList     *file_list.FileList
	gridView           *gtk.GridView
	stack              *gtk.Stack
	store              *gio.ListStore
	folderIcon         *gtk.Image
	folderName         *gtk.Label
	rightBox           *gtk.Box
	specialPathManager *special_path.SpecialPathManager
}

func NewImageDirPreviewer(path string, changePath func(string), specialPathManager *special_path.SpecialPathManager) *ImageDirPreviewer {
	viewer := &ImageDirPreviewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 2),
		Path:               path,
		SortOrder:          SortByName,
		changePath:         changePath,
		FileViewerList:     file_list.NewFileList(false, specialPathManager, nil),
		stack:              gtk.NewStack(),
		specialPathManager: specialPathManager,
	}
	viewer.folderIcon = gtk.NewImageFromIconName("folder-symbolic")
	viewer.folderIcon.SetPixelSize(20)
	viewer.folderName = gtk.NewLabel("")
	viewer.folderName.AddCSSClass("dir-previewer-title")
	viewer.folderName.SetEllipsize(1) // pango.EllipsizeEnd

	viewer.FileViewerList.SetHExpand(true)
	viewer.FileViewerList.SetVExpand(true)
	viewer.Box.SetVExpand(true)
	viewer.Box.SetHExpand(true)

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.AddCSSClass("dir-previewer-header")
	viewer.Box.Append(headerBox)

	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	leftBox.SetHExpand(true)
	headerBox.Append(leftBox)

	leftBox.Append(viewer.folderIcon)
	leftBox.Append(viewer.folderName)

	rightBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.Append(rightBox)
	viewer.rightBox = rightBox

	gridViewButton := gtk.NewButtonFromIconName("view-list-symbolic")
	rightBox.Append(gridViewButton)

	viewer.FileViewerList.PathChanged = func(path string) {
		changePath(path)
	}

	viewer.store = gio.NewListStore(glib.TypeObject)
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		box := gtk.NewBox(gtk.OrientationVertical, 6)
		box.AddCSSClass("preview-grid-item")
		box.SetHExpand(true)
		box.SetVExpand(true)
		box.SetSizeRequest(80, 80)
		image := gtk.NewImage()
		image.SetPixelSize(48)
		label := gtk.NewLabel("")
		label.SetWrap(true)
		label.SetMaxWidthChars(12)
		label.SetEllipsize(1) // pango.EllipsizeEnd
		box.Append(image)
		box.Append(label)
		item.SetChild(box)
	})
	factory.ConnectBind(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		box := item.Child().(*gtk.Box)
		image := box.FirstChild().(*gtk.Image)
		label := box.LastChild().(*gtk.Label)
		obj := item.Item()
		str := obj.Cast().(*gtk.StringObject).String()
		label.SetText(str)
		fullPath := filepath.Join(viewer.Path, str)
		
		pixbuf, err := thumbnail.Generate(fullPath)
		if err == nil && pixbuf != nil {
			image.SetFromPaintable(pixbuf)
		} else {
			image.SetFromIconName(fileops.GetIconForFile(str))
		}
	})

	viewer.gridView = gtk.NewGridView(gtk.NewSingleSelection(viewer.store), &factory.ListItemFactory)
	viewer.gridView.SetMaxColumns(6)
	viewer.gridView.SetMinColumns(2)

	viewer.gridView.ConnectActivate(func(position uint) {
		if int(position) < len(viewer.FileViewerList.Items) {
			item := viewer.FileViewerList.Items[position]
			if item.IsDir {
				changePath(item.Path)
			}
		}
	})

	gridScrolled := gtk.NewScrolledWindow()
	gridScrolled.SetChild(viewer.gridView)
	gridScrolled.SetHExpand(true)
	gridScrolled.SetVExpand(true)

	viewer.stack.SetVExpand(true)
	viewer.stack.SetHExpand(true)
	viewer.stack.AddTitled(viewer.FileViewerList, "list", "List")
	viewer.stack.AddTitled(gridScrolled, "grid", "Grid")

	emptyLabel := gtk.NewLabel("No Images Found")
	emptyLabel.AddCSSClass("preview-title")
	emptyLabel.SetHAlign(gtk.AlignCenter)
	emptyLabel.SetVAlign(gtk.AlignCenter)
	viewer.stack.AddTitled(emptyLabel, "empty", "Empty")

	viewer.stack.SetVisibleChildName("grid")
	viewer.Box.Append(viewer.stack)

	gridViewButton.ConnectClicked(func() {
		if viewer.stack.VisibleChildName() == "list" {
			viewer.stack.SetVisibleChildName("grid")
			gridViewButton.SetIconName("view-list-symbolic")
		} else {
			viewer.stack.SetVisibleChildName("list")
			gridViewButton.SetIconName("view-grid-symbolic")
		}
	})

	viewer.Refresh()

	return viewer
}

func (viewer *ImageDirPreviewer) SetPath(path string) {
	viewer.Path = path
	viewer.folderName.SetText(filepath.Base(path))
	viewer.Refresh()
}

func (viewer *ImageDirPreviewer) Refresh() {
	if viewer.Path == "" {
		return
	}

	var items []*types.ListItem
	specialPath := viewer.specialPathManager.GetPath(viewer.Path)
	if specialPath != nil {
		allPaths := specialPath.GetItems()
		for _, item := range allPaths {
			if fileops.IsImage(item.Name) {
				items = append(items, item)
			}
		}
	} else {
		entries, err := os.ReadDir(viewer.Path)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		var filteredEntries []os.DirEntry
		for _, entry := range entries {
			if !entry.IsDir() && fileops.IsImage(entry.Name()) {
				filteredEntries = append(filteredEntries, entry)
			}
		}

		sort.Slice(filteredEntries, func(i, j int) bool {
			switch viewer.SortOrder {
			case SortByTime:
				infoI, errI := filteredEntries[i].Info()
				infoJ, errJ := filteredEntries[j].Info()
				if errI != nil || errJ != nil {
					return false
				}
				return infoI.ModTime().After(infoJ.ModTime())
			default:
				return strings.Title(filteredEntries[i].Name()) < strings.Title(filteredEntries[j].Name())
			}
		})

		for _, entry := range filteredEntries {
			fullPath := filepath.Join(viewer.Path, entry.Name())
			var group string
			if viewer.SortOrder == SortByTime {
				info, err := entry.Info()
				if err != nil {
					group = "Unknown"
				} else {
					group = fileops.GetGroupForTime(info.ModTime())
				}
			} else {
				name := entry.Name()
				runes := []rune(strings.Title(name))
				firstRune := runes[0]
				group = string(firstRune)
			}
			listItem := &types.ListItem{
				Name:  entry.Name(),
				Path:  fullPath,
				Group: group,
				IsDir: entry.IsDir(),
			}
			info, err := entry.Info()
			if err == nil {
				listItem.Size = info.Size()
			}
			items = append(items, listItem)
		}
	}

	viewer.store.RemoveAll()
	for _, item := range items {
		viewer.store.Append(gtk.NewStringObject(item.Name).Object)
	}
	viewer.FileViewerList.SetItems(items)
	viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(viewer.Path))

	if len(items) == 0 {
		viewer.stack.SetVisibleChildName("empty")
		viewer.rightBox.SetVisible(false)
	} else {
		viewer.rightBox.SetVisible(true)
		if viewer.stack.VisibleChildName() == "empty" {
			viewer.stack.SetVisibleChildName("grid")
		}
	}
}
