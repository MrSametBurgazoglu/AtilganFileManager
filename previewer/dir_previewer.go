package previewer

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

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

type FileType int

const (
	TypeDir FileType = iota
	TypeExec
	TypeHidden
	TypeTemp
	TypeOther
	TypeDoc
	TypeMedia
)

type DirPreviewer struct {
	*gtk.Box
	Path               string
	SortOrder          SortOrder
	Filters            []string
	DefaultFilters     []string
	FiltersMap         map[string]bool
	popover            *gtk.Popover
	changePath         func(string)
	FileViewerList     *file_list.FileList
	gridView           *gtk.GridView
	stack              *gtk.Stack
	store              *gio.ListStore
	folderIcon         *gtk.Image
	folderName         *gtk.Label
	specialPathManager *special_path.SpecialPathManager
}

func NewDirPreviewer(path string, changePath func(string), specialPathManager *special_path.SpecialPathManager) *DirPreviewer {
	viewer := &DirPreviewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 6),
		Path:               path,
		SortOrder:          SortByName,
		FiltersMap:         make(map[string]bool),
		changePath:         changePath,
		FileViewerList:     file_list.NewFileList(false, nil, nil),
		stack:              gtk.NewStack(),
		DefaultFilters:     []string{"Directories", "Executables", "Hidden"},
		folderIcon:         gtk.NewImageFromIconName("folder-symbolic"),
		folderName:         gtk.NewLabel(""),
		specialPathManager: specialPathManager,
	}
	//viewer.Box.SetVExpand(true)

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.AddCSSClass("dir-previewer-header")
	viewer.Box.Append(headerBox)

	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	leftBox.SetHExpand(true)
	headerBox.Append(leftBox)

	leftBox.Append(viewer.folderIcon)
	leftBox.Append(viewer.folderName)

	rightBox := gtk.NewBox(gtk.OrientationHorizontal, 0)
	headerBox.Append(rightBox)

	sortButton := gtk.NewButtonFromIconName("view-sort-descending-symbolic")

	gridViewButton := gtk.NewButtonFromIconName("view-grid-symbolic")
	rightBox.Append(gridViewButton)
	rightBox.Append(sortButton)

	filterButton := gtk.NewMenuButton()
	filterButton.SetIconName("preferences-system-symbolic")
	rightBox.Append(filterButton)

	viewer.popover = gtk.NewPopover()
	filterButton.SetPopover(viewer.popover)

	popoverBox := gtk.NewBox(gtk.OrientationVertical, 6)
	viewer.popover.SetChild(popoverBox)

	viewer.FileViewerList.SelectionChanged = func(index int) {
		selectedItem := viewer.FileViewerList.Items[index]
		if selectedItem.IsDir {
			changePath(selectedItem.Path)
		}
	}

	viewer.store = gio.NewListStore(glib.TypeObject)
	factory := gtk.NewSignalListItemFactory()
	factory.ConnectSetup(func(o *glib.Object) {
		item := o.Cast().(*gtk.ListItem)
		box := gtk.NewBox(gtk.OrientationVertical, 6)
		box.SetHExpand(false)
		image := gtk.NewImage()
		image.SetPixelSize(64)
		label := gtk.NewLabel("")
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
		info, err := os.Stat(fullPath)
		if err == nil {
			if info.IsDir() {
				image.SetFromIconName(fileops.GetIconForFolderSymbolic(fullPath))
			} else {
				pixbuf, err := thumbnail.Generate(fullPath)
				if err == nil && pixbuf != nil {
					image.SetFromPaintable(pixbuf)
				} else {
					image.SetFromIconName("text-x-generic-symbolic")
				}
			}
		}
	})

	viewer.gridView = gtk.NewGridView(gtk.NewSingleSelection(viewer.store), &factory.ListItemFactory)
	viewer.gridView.SetVisible(false)
	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(viewer.gridView)
	scrolled.SetMaxContentHeight(600)
	scrolled.SetMaxContentWidth(500)
	scrolled.SetHExpand(false)
	scrolled.SetVExpand(false)
	viewer.stack.AddTitled(viewer.FileViewerList, "list", "List")
	viewer.stack.AddTitled(scrolled, "grid", "Grid")
	viewer.stack.SetVisibleChildName("list")
	viewer.Box.Append(viewer.stack)

	gridViewButton.ConnectClicked(func() {
		if viewer.stack.VisibleChildName() == "list" {
			viewer.gridView.SetVisible(true)
			viewer.stack.SetVisibleChildName("grid")
			gridViewButton.SetIconName("view-list-symbolic")
		} else {
			viewer.stack.SetVisibleChildName("list")
			gridViewButton.SetIconName("view-grid-symbolic")
		}
	})

	sortButton.ConnectClicked(func() {
		if viewer.SortOrder == SortByTime {
			viewer.SortOrder = SortByName
		} else {
			viewer.SortOrder = SortByTime
		}
		viewer.Refresh(false)
	})

	viewer.Refresh(false)

	return viewer
}

func (viewer *DirPreviewer) SetPath(path string) {
	viewer.Path = path
	viewer.folderName.SetText(filepath.Base(path))
	viewer.Refresh(true)
}

func (viewer *DirPreviewer) Refresh(newFilter bool) {
	if viewer.Path == "" {
		return
	}

	specialPath := viewer.specialPathManager.GetPath(viewer.Path)
	if specialPath != nil {
		items := specialPath.GetItems()
		viewer.FileViewerList.SetItems(items)
		viewer.folderName.SetText(specialPath.GetName())
		viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(viewer.Path))
		return
	}

	entries, err := os.ReadDir(viewer.Path)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	if newFilter {
		viewer.Filters = []string{}
		viewer.FiltersMap = make(map[string]bool)
		viewer.DefaultFilters = make([]string, 0)
		extensions := []string{}
		hasDir := false
		hasExec := false
		hasHidden := false
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext != "" {
					isExist := viewer.FiltersMap[ext]
					viewer.FiltersMap[ext] = true
					if !isExist {
						extensions = append(extensions, ext)
					}
				} else {
					hasExec = true
				}
			} else if strings.HasPrefix(entry.Name(), ".") {
				hasHidden = true
			} else {
				hasDir = true
			}
		}
		if hasDir {
			viewer.FiltersMap["Directories"] = true
			viewer.DefaultFilters = append(viewer.DefaultFilters, "Directories")
		}
		if hasExec {
			viewer.FiltersMap["Executables"] = true
			viewer.DefaultFilters = append(viewer.DefaultFilters, "Executables")
		}
		if hasHidden {
			viewer.FiltersMap["Hidden"] = false
			viewer.DefaultFilters = append(viewer.DefaultFilters, "Hidden")
		}
		sort.Strings(extensions)
		viewer.Filters = append(viewer.Filters, extensions...)
		viewer.UpdateFilterPopover()
	}

	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		fileType := getFileType(entry)
		show := false
		if fileType == TypeDir {
			if viewer.FiltersMap["Directories"] {
				show = true
			}
		} else if fileType == TypeExec {
			if viewer.FiltersMap["Executables"] {
				show = true
			}
		} else if fileType == TypeHidden {
			if viewer.FiltersMap["Hidden"] {
				show = true
			}
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if viewer.FiltersMap[ext] {
				show = true
			}
		}

		if !show {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
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

	newFiles := make([]*types.ListItem, 0)
	viewer.store.RemoveAll()
	for _, entry := range filteredEntries {
		fullPath := filepath.Join(viewer.Path, entry.Name())
		var group string
		if viewer.SortOrder == SortByTime {
			info, err := entry.Info()
			if err != nil {
				group = "Unknown"
			} else {
				group = getGroupForTime(info.ModTime())
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
		if listItem.IsDir {
			listItem.ItemCount = getDirItemCount(fullPath)
		} else {
			info, err := entry.Info()
			if err == nil {
				listItem.Size = info.Size()
			}

		}
		newFiles = append(newFiles, listItem)
		viewer.store.Append(gtk.NewStringObject(entry.Name()).Object)
	}
	viewer.FileViewerList.SetItems(newFiles)
	viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(viewer.Path))
}

func (viewer *DirPreviewer) UpdateFilterPopover() {
	popoverBox := viewer.popover.Child().(*gtk.Box)
	for child := popoverBox.FirstChild(); child != nil; child = popoverBox.FirstChild() {
		popoverBox.Remove(child)
	}

	for _, filter := range viewer.DefaultFilters {
		checkButton := gtk.NewCheckButtonWithLabel(filter)
		checkButton.SetActive(viewer.FiltersMap[filter])
		filterName := filter
		checkButton.ConnectToggled(func() {
			viewer.FiltersMap[filterName] = checkButton.Active()
			viewer.UpdateFilterPopover()
			viewer.Refresh(false)
		})
		popoverBox.Append(checkButton)
	}
	seperator := gtk.NewSeparator(gtk.OrientationHorizontal)
	popoverBox.Append(seperator)

	for _, filter := range viewer.Filters {
		checkButton := gtk.NewCheckButtonWithLabel(filter)
		checkButton.SetActive(viewer.FiltersMap[filter])
		filterName := filter
		checkButton.ConnectToggled(func() {
			viewer.FiltersMap[filterName] = checkButton.Active()
			viewer.UpdateFilterPopover()
			viewer.Refresh(false)
		})
		popoverBox.Append(checkButton)
	}
}

func isImage(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".png") ||
		strings.HasSuffix(fileName, ".jpg") ||
		strings.HasSuffix(fileName, ".jpeg") ||
		strings.HasSuffix(fileName, ".gif")
}

func getGroupForTime(modTime time.Time) string {
	now := time.Now()
	duration := now.Sub(modTime)

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if modTime.After(todayStart) {
		return "Today"
	}

	if duration.Hours() <= 24 {
		return "Last 24 hours"
	}

	if duration.Hours() <= 24*7 {
		return "Last Week"
	}

	if duration.Hours() <= 24*30 {
		return "Last Month"
	}

	return "Later"
}

func getDirItemCount(dirPath string) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}
	return len(entries)
}

func getFileType(entry os.DirEntry) FileType {
	fileName := entry.Name()
	if strings.HasPrefix(fileName, ".") {
		return TypeHidden
	}

	if strings.HasPrefix(fileName, "~") {
		return TypeTemp
	}

	if entry.IsDir() {
		return TypeDir
	}

	info, err := entry.Info()
	if err != nil {
		return TypeOther
	}

	if info.Mode()&0111 != 0 {
		return TypeExec
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".doc", ".docx", ".pdf", ".txt", ".md":
		return TypeDoc
	case ".jpg", ".jpeg", ".png", ".gif", ".mp3", ".mp4", ".avi", ".mkv":
		return TypeMedia
	}

	return TypeOther
}
