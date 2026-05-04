package viewer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/file_list"
	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/git"
	"github.com/MrSametBurgazoglu/atilgan/special_path"
	"github.com/MrSametBurgazoglu/atilgan/types"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

type FileViewHistory struct {
	Path  string
	Index int
}

type FileViewer struct {
	*gtk.Box
	Path               string
	SortOrder          types.SortOrder
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
	FileViewerList     *file_list.FileGrid
	GitManager         *git.GitManager
	stack              *gtk.Stack
	specialPathManager *special_path.SpecialPathManager
}

func NewFileViewer(mainWindow *gtk.Window, path string, pathChanged func(string), specialPathManager *special_path.SpecialPathManager) *FileViewer {
	viewer := &FileViewer{
		Box:                gtk.NewBox(gtk.OrientationVertical, 6),
		Path:               path,
		SortOrder:          types.SortByName,
		SearchValue:        "",
		FiltersMap:         make(map[string]bool),
		FileViewerHistory:  make(map[string]*FileViewHistory),
		FileViewerList:     file_list.NewFileGrid(true, specialPathManager, mainWindow),
		DefaultFilters:     []string{"Directories", "Executables", "Hidden"},
		GitManager:         git.NewGitManager(),
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

	// Refresh Git Status
	viewer.GitManager.Refresh(viewer.Path)

	if newFilter {
		viewer.Filters = []string{}
		viewer.FiltersMap = make(map[string]bool)
		viewer.DefaultFilters = make([]string, 0)
		hasDir := false
		hasExec := false
		hasHidden := false
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				hasExec = true
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
		viewer.UpdateFilterPopover()
	}

	var filteredEntries []os.DirEntry
	for _, entry := range entries {
		if viewer.SearchValue != "" {
			if !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(viewer.SearchValue)) {
				continue
			}
		}
		fileType := fileops.GetFileType(entry)
		show := false
		if fileType == types.TypeDir {
			if viewer.FiltersMap["Directories"] {
				show = true
			}
		} else if fileType == types.TypeExec {
			if viewer.FiltersMap["Executables"] {
				show = true
			}
		} else if fileType == types.TypeHidden {
			if viewer.FiltersMap["Hidden"] {
				show = true
			}
		} else {
			show = true
		}

		if !show {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	sort.Slice(filteredEntries, func(i, j int) bool {
		switch viewer.SortOrder {
		case types.SortByTime:
			infoI, errI := filteredEntries[i].Info()
			infoJ, errJ := filteredEntries[j].Info()
			if errI != nil || errJ != nil {
				return false
			}
			return infoI.ModTime().After(infoJ.ModTime())
		case types.SortBySize:
			isDirI := filteredEntries[i].IsDir()
			isDirJ := filteredEntries[j].IsDir()
			if isDirI && !isDirJ {
				return true
			}
			if !isDirI && isDirJ {
				return false
			}
			infoI, errI := filteredEntries[i].Info()
			infoJ, errJ := filteredEntries[j].Info()
			if errI == nil && errJ == nil {
				if infoI.Size() != infoJ.Size() {
					return infoI.Size() > infoJ.Size()
				}
			}
			return strings.ToLower(filteredEntries[i].Name()) < strings.ToLower(filteredEntries[j].Name())
		case types.SortByType:
			typeI := fileops.GetFileType(filteredEntries[i])
			typeJ := fileops.GetFileType(filteredEntries[j])
			if typeI != typeJ {
				return typeI < typeJ
			}
			return strings.ToLower(filteredEntries[i].Name()) < strings.ToLower(filteredEntries[j].Name())
		default:
			return strings.ToLower(filteredEntries[i].Name()) < strings.ToLower(filteredEntries[j].Name())
		}
	})

	newFiles := make([]*types.ListItem, 0)
	var sizeThresholds []int64
	if viewer.SortOrder == types.SortBySize {
		sizeThresholds = fileops.CalculateSizeThresholds(filteredEntries)
	}

	for _, entry := range filteredEntries {
		fullPath := filepath.Join(viewer.Path, entry.Name())
		var group string
		if viewer.SortOrder == types.SortByTime {
			info, err := entry.Info()
			if err != nil {
				group = "Unknown"
			} else {
				group = fileops.GetGroupForTime(info.ModTime())
			}
		} else if viewer.SortOrder == types.SortBySize {
			if entry.IsDir() {
				group = "Directories"
			} else {
				info, _ := entry.Info()
				group = fileops.GetGroupForSize(info.Size(), sizeThresholds)
			}
		} else if viewer.SortOrder == types.SortByType {
			group = fileops.GetGroupForType(entry)
		} else {
			name := entry.Name()
			runes := []rune(strings.ToLower(name))
			firstRune := runes[0]
			group = string(firstRune)
		}
		listItem := &types.ListItem{
			Name:      entry.Name(),
			Path:      fullPath,
			Group:     group,
			IsDir:     entry.IsDir(),
			GitStatus: string(viewer.GitManager.GetStatus(fullPath)),
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

	defaultGrid := gtk.NewGrid()
	defaultGrid.SetColumnSpacing(12)
	defaultGrid.SetRowSpacing(6)
	popoverBox.Append(defaultGrid)

	for i, filter := range viewer.DefaultFilters {
		checkButton := gtk.NewCheckButtonWithLabel(filter)
		checkButton.SetActive(viewer.FiltersMap[filter])
		filterName := filter
		checkButton.ConnectToggled(func() {
			viewer.FiltersMap[filterName] = checkButton.Active()
			viewer.UpdateFilterPopover()
			viewer.Refresh(false)
		})
		defaultGrid.Attach(checkButton, i%2, i/2, 1, 1)
	}

	if len(viewer.Filters) > 0 {
		seperator := gtk.NewSeparator(gtk.OrientationHorizontal)
		popoverBox.Append(seperator)

		filterGrid := gtk.NewGrid()
		filterGrid.SetColumnSpacing(12)
		filterGrid.SetRowSpacing(6)
		popoverBox.Append(filterGrid)

		for i, filter := range viewer.Filters {
			checkButton := gtk.NewCheckButtonWithLabel(filter)
			checkButton.SetActive(viewer.FiltersMap[filter])
			filterName := filter
			checkButton.ConnectToggled(func() {
				viewer.FiltersMap[filterName] = checkButton.Active()
				viewer.UpdateFilterPopover()
				viewer.Refresh(false)
			})
			filterGrid.Attach(checkButton, i%2, i/2, 1, 1)
		}
	}
}

func getDirItemCount(dirPath string) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}

	return len(entries)
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

func (viewer *FileViewer) AddSelectedToCopyCut() {
	for idx, selected := range viewer.FileViewerList.SelectedIdxs {
		if selected {
			viewer.AddCopyCutItem(idx)
		}
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
		errors := fileops.CutFiles(filePaths, viewer.Path, progress)
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
