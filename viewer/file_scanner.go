package viewer

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/fileops"
	"github.com/MrSametBurgazoglu/atilgan/git"
	"github.com/MrSametBurgazoglu/atilgan/types"
)

type FileScanner struct {
	Path        string
	SortOrder   types.SortOrder
	SearchValue string
	FiltersMap  map[string]bool
	GitManager  *git.GitManager
}

func NewFileScanner(path string, sortOrder types.SortOrder, gitManager *git.GitManager) *FileScanner {
	return &FileScanner{
		Path:        path,
		SortOrder:   sortOrder,
		FiltersMap:  make(map[string]bool),
		GitManager:  gitManager,
	}
}

func (s *FileScanner) Scan() ([]*types.ListItem, error) {
	entries, err := os.ReadDir(s.Path)
	if err != nil {
		return nil, err
	}

	s.GitManager.Refresh(s.Path)

	filteredEntries := s.filterEntries(entries)
	s.sortEntries(filteredEntries)

	var sizeThresholds []int64
	if s.SortOrder == types.SortBySize {
		sizeThresholds = fileops.CalculateSizeThresholds(filteredEntries)
	}

	items := make([]*types.ListItem, 0, len(filteredEntries))
	for _, entry := range filteredEntries {
		fullPath := filepath.Join(s.Path, entry.Name())
		items = append(items, s.createListItem(entry, fullPath, sizeThresholds))
	}

	return items, nil
}

func (s *FileScanner) filterEntries(entries []os.DirEntry) []os.DirEntry {
	filtered := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if s.SearchValue != "" && !strings.HasPrefix(strings.ToLower(entry.Name()), strings.ToLower(s.SearchValue)) {
			continue
		}

		fileType := fileops.GetFileType(entry)
		show := false
		switch fileType {
		case types.TypeDir:
			show = s.FiltersMap["Directories"]
		case types.TypeExec:
			show = s.FiltersMap["Executables"]
		case types.TypeHidden:
			show = s.FiltersMap["Hidden"]
		default:
			show = true
		}

		if show {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (s *FileScanner) sortEntries(entries []os.DirEntry) {
	sort.Slice(entries, func(i, j int) bool {
		switch s.SortOrder {
		case types.SortByTime:
			infoI, _ := entries[i].Info()
			infoJ, _ := entries[j].Info()
			if infoI == nil || infoJ == nil {
				return false
			}
			return infoI.ModTime().After(infoJ.ModTime())
		case types.SortBySize:
			isDirI := entries[i].IsDir()
			isDirJ := entries[j].IsDir()
			if isDirI && !isDirJ {
				return true
			}
			if !isDirI && isDirJ {
				return false
			}
			infoI, _ := entries[i].Info()
			infoJ, _ := entries[j].Info()
			if infoI != nil && infoJ != nil && infoI.Size() != infoJ.Size() {
				return infoI.Size() > infoJ.Size()
			}
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		case types.SortByType:
			typeI := fileops.GetFileType(entries[i])
			typeJ := fileops.GetFileType(entries[j])
			if typeI != typeJ {
				return typeI < typeJ
			}
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		default:
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		}
	})
}

func (s *FileScanner) createListItem(entry os.DirEntry, fullPath string, sizeThresholds []int64) *types.ListItem {
	var group string
	switch s.SortOrder {
	case types.SortByTime:
		info, _ := entry.Info()
		if info != nil {
			group = fileops.GetGroupForTime(info.ModTime())
		} else {
			group = "Unknown"
		}
	case types.SortBySize:
		if entry.IsDir() {
			group = "Directories"
		} else {
			info, _ := entry.Info()
			if info != nil {
				group = fileops.GetGroupForSize(info.Size(), sizeThresholds)
			}
		}
	case types.SortByType:
		group = fileops.GetGroupForType(entry)
	default:
		runes := []rune(strings.ToLower(entry.Name()))
		if len(runes) > 0 {
			group = string(runes[0])
		}
	}

	item := &types.ListItem{
		Name:      entry.Name(),
		Path:      fullPath,
		Group:     group,
		IsDir:     entry.IsDir(),
		GitStatus: string(s.GitManager.GetStatus(fullPath)),
	}

	if item.IsDir {
		item.ItemCount = fileops.GetDirItemCount(fullPath)
	} else {
		info, _ := entry.Info()
		if info != nil {
			item.Size = info.Size()
		}
	}
	return item
}
