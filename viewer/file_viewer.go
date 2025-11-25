package viewer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MrSametBurgazoglu/atilgan/create_popup"
	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
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

type FileViewHistory struct {
	Path  string
	Index int
}

type FileViewer struct {
	*gtk.Box
	Path               string
	SortOrder          SortOrder
	SearchValue        string
	SearchRevealer     *gtk.Revealer
	SearchEntry        *gtk.SearchEntry
	Filters            []string
	DefaultFilters     []string
	CopiedCuttedFiles  []string
	FiltersMap         map[string]bool
	IsCopy             bool
	IsCut              bool
	folderIcon         *gtk.Image
	folderName         *gtk.Label
	popover            *gtk.Popover
	createPopover      *create_popup.CreatePopover
	FileViewerHistory  map[string]*FileViewHistory
	FileViewerList     *file_list.FileList
	specialPathManager *special_path.SpecialPathManager
}

func NewFileViewer(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager) *FileViewer {
	viewer := &FileViewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 6),
		Path:               path,
		SortOrder:          SortByName,
		SearchValue:        "",
		FiltersMap:         make(map[string]bool),
		FileViewerHistory:  make(map[string]*FileViewHistory),
		FileViewerList:     file_list.NewFileList(true, specialPathManager, mainWindow),
		DefaultFilters:     []string{"Directories", "Executables", "Hidden"},
		folderIcon:         gtk.NewImageFromIconName("folder-symbolic"),
		folderName:         gtk.NewLabel(filepath.Base(path)),
		specialPathManager: specialPathManager,
	}
	viewer.SetVExpand(true)

	headerBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	viewer.Box.Append(headerBox)

	separator := gtk.NewSeparator(gtk.OrientationHorizontal)
	viewer.Box.Append(separator)

	searchBar := gtk.NewBox(gtk.OrientationHorizontal, 6)
	searchBar.SetHExpand(true)
	viewer.SearchEntry = gtk.NewSearchEntry()
	viewer.SearchEntry.SetHExpand(true)
	searchBar.Append(viewer.SearchEntry)
	searchCloseButton := gtk.NewButtonFromIconName("window-close-symbolic")
	searchBar.Append(searchCloseButton)

	viewer.SearchRevealer = gtk.NewRevealer()
	viewer.SearchRevealer.SetVisible(false)
	viewer.SearchRevealer.SetChild(searchBar)
	viewer.SearchRevealer.SetTransitionType(gtk.RevealerTransitionTypeSlideLeft)
	viewer.Box.Append(viewer.SearchRevealer)

	viewer.SearchEntry.ConnectSearchChanged(func() {
		viewer.SearchValue = viewer.SearchEntry.Text()
		viewer.Refresh(false)
	})

	searchCloseButton.ConnectClicked(func() {
		viewer.SearchRevealer.SetRevealChild(false)
	})

	leftBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	leftBox.SetHExpand(true)
	headerBox.Append(leftBox)
	leftBox.Append(viewer.folderIcon)
	leftBox.Append(viewer.folderName)

	rightBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	headerBox.Append(rightBox)

	sortButton := gtk.NewButtonFromIconName("view-sort-descending-symbolic")

	newButton := gtk.NewMenuButton()
	newButton.SetIconName("list-add-symbolic")
	createPopover := create_popup.NewCreatePopover(mainWindow, pathChanged)
	viewer.createPopover = createPopover
	newButton.SetPopover(createPopover)

	terminalButton := gtk.NewButtonFromIconName("utilities-terminal-symbolic")
	terminalButton.ConnectClicked(func() {
		// macos cmd := exec.Command("open", "-a", "Terminal", viewer.Path)

		cmd := exec.Command("x-terminal-emulator", "-d", viewer.Path)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Start()
		if err != nil {
			fmt.Printf("Error opening Terminal: %v\n", err)
		}
	})

	filterButton := gtk.NewMenuButton()
	filterButton.SetIconName("preferences-system-symbolic")
	rightBox.Append(newButton)
	rightBox.Append(terminalButton)
	rightBox.Append(sortButton)
	rightBox.Append(filterButton)

	viewer.popover = gtk.NewPopover()
	filterButton.SetPopover(viewer.popover)

	popoverBox := gtk.NewBox(gtk.OrientationVertical, 6)
	viewer.popover.SetChild(popoverBox)
	viewer.Box.Append(viewer.FileViewerList)

	sortButton.ConnectClicked(func() {
		if viewer.SortOrder == SortByTime {
			viewer.SortOrder = SortByName
		} else {
			viewer.SortOrder = SortByTime
		}
		viewer.Refresh(false)
	})

	viewer.Refresh(true)

	return viewer
}

func (viewer *FileViewer) SetPath(path string) {
	viewer.Path = path
	viewer.createPopover.CurrentPath = path
	viewer.folderName.SetText(filepath.Base(path))
	viewer.Refresh(true)
}

func (viewer *FileViewer) SetFolderName(name string) {
	viewer.folderName.SetText(name)
	viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(name))
}

func (viewer *FileViewer) Refresh(newFilter bool) {
	if viewer.Path == "" {
		return
	}
	specialPath := viewer.specialPathManager.GetPath(viewer.Path)
	if specialPath != nil {
		items := specialPath.GetItems()
		viewer.FileViewerList.SetItems(items)
		viewer.SetFolderName(specialPath.GetName())
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
		if viewer.SearchValue != "" {
			if !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(viewer.SearchValue)) {
				continue
			}
		}
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
	for _, entry := range filteredEntries {
		fullPath := path.Join(viewer.Path, entry.Name())
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
	}
	viewer.FileViewerList.SetItems(newFiles)
	println("this is where I set folder icon")
	viewer.folderIcon.SetFromIconName(fileops.GetIconForFolderSymbolic(viewer.Path))
}

func (viewer *FileViewer) UpdateFilterPopover() {
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

func (viewer *FileViewer) CleanCopyCutFiles() {
	viewer.CopiedCuttedFiles = []string{}
	viewer.FileViewerList.CleanCopyCutItems()
}

func (viewer *FileViewer) AddCopyCutItem(index int) {
	item := viewer.FileViewerList.Items[index]
	if viewer.FileViewerList.AddCopyCutItem(item.Path) {
		viewer.CopiedCuttedFiles = append(viewer.CopiedCuttedFiles, item.Path)
	}
}

func (viewer *FileViewer) ExecuteCopyPaste(progress func(float64)) error {
	if !viewer.IsCopy {
		return errors.New("not in copy mode")
	}
	if len(viewer.CopiedCuttedFiles) == 0 {
		return errors.New("no files to copy")
	}
	if viewer.Path == "" {
		return errors.New("no destination path")
	}
	filePaths := make([]string, len(viewer.CopiedCuttedFiles))

	for i, file := range viewer.CopiedCuttedFiles {
		filePaths[i] = file
	}

	if viewer.IsCut {
		errors := fileops.CutFiles(filePaths, viewer.Path)
		if errors != nil && len(errors) > 0 {
			return errors[0]
		}
	} else {
		errors := fileops.CopyFiles(filePaths, viewer.Path, progress)
		if errors != nil && len(errors) > 0 {
			println(errors[0].Error())
			return errors[0]
		}
	}

	return nil
}
