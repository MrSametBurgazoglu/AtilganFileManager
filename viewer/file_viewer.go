package viewer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

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
	FilterBox          *gtk.Box
	FileViewerHistory  map[string]*FileViewHistory
	FileViewerList     *file_list.FileList
	stack              *gtk.Stack
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
		specialPathManager: specialPathManager,
		FilterBox:          gtk.NewBox(gtk.OrientationVertical, 6),
	}
	viewer.SetVExpand(true)
	viewer.SetHExpand(true)

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

	viewer.stack = gtk.NewStack()
	viewer.stack.SetVExpand(true)
	viewer.stack.SetHExpand(true)
	viewer.stack.AddTitled(viewer.FileViewerList, "list", "List")

	emptyLabel := gtk.NewLabel("Empty Directory")
	emptyLabel.AddCSSClass("preview-title")
	emptyLabel.SetHAlign(gtk.AlignCenter)
	emptyLabel.SetVAlign(gtk.AlignCenter)
	viewer.stack.AddTitled(emptyLabel, "empty", "Empty")

	viewer.Box.Append(viewer.stack)

	viewer.Refresh(true)

	return viewer
}

func (viewer *FileViewer) SetPath(path string) {
	viewer.Path = path
	viewer.Refresh(true)
}

func (viewer *FileViewer) SetFolderName(name string) {
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
		if len(items) == 0 {
			viewer.stack.SetVisibleChildName("empty")
		} else {
			viewer.stack.SetVisibleChildName("list")
		}
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

	if len(newFiles) == 0 {
		viewer.stack.SetVisibleChildName("empty")
	} else {
		viewer.stack.SetVisibleChildName("list")
	}
}

func (viewer *FileViewer) UpdateFilterPopover() {
	popoverBox := viewer.FilterBox
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

func (viewer *FileViewer) OpenTerminal() {
	if viewer.Path == "" {
		return
	}

	// Check if it's a real directory
	info, err := os.Stat(viewer.Path)
	if err != nil || !info.IsDir() {
		return
	}

	if runtime.GOOS == "darwin" {
		exec.Command("open", "-a", "Terminal", viewer.Path).Start()
		return
	}

	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "start", "cmd.exe", "/K", "cd /d", viewer.Path).Start()
		return
	}

	// Linux: Try $TERMINAL then common emulators
	terminal := os.Getenv("TERMINAL")
	if terminal == "" {
		terminals := []string{"ptyxis", "gnome-console", "kgx", "xdg-terminal-exec", "x-terminal-emulator", "gnome-terminal", "konsole", "xfce4-terminal", "alacritty", "kitty", "terminator", "lxterminal", "mate-terminal", "xterm"}
		for _, t := range terminals {
			if _, err := exec.LookPath(t); err == nil {
				terminal = t
				break
			}
		}
	}

	if terminal != "" {
		cmd := exec.Command(terminal)
		cmd.Dir = viewer.Path
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Failed to start terminal %s: %v\n", terminal, err)
		}
		return
	}

	fmt.Printf("No terminal emulator found for path: %s\n", viewer.Path)
}
