package recent

import (
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/types"
)

type RecentPath struct {
	path          string
	recentManager *RecentManager
}

func NewRecentPath(recentManager *RecentManager) *RecentPath {
	return &RecentPath{
		path:          "recent://",
		recentManager: recentManager,
	}
}

func (r *RecentPath) GetItems() []*types.ListItem {
	var items []*types.ListItem
	paths := r.recentManager.GetPaths()
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		listItem := &types.ListItem{
			Name:  filepath.Base(path),
			IsDir: info.IsDir(),
			Path:  path,
		}
		if listItem.IsDir {
			listItem.ItemCount = getDirItemCount(path)
		} else {
			listItem.Size = info.Size()
		}
		items = append(items, listItem)
	}
	return items
}

func (r *RecentPath) GetPath() string {
	return r.path
}

func (r *RecentPath) GetParentPath() string {
	return ""
}

func (r *RecentPath) GetName() string {
	return "Recent"
}

func getDirItemCount(dirPath string) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}
	return len(entries)
}
