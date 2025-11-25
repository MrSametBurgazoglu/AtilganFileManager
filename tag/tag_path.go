package tag

import (
	"os"
	"path/filepath"

	"github.com/MrSametBurgazoglu/atilgan/types"
)

type TagPath struct {
	tag        string
	tagManager *TagManager
}

func NewTagPath(tag string, tagManager *TagManager) *TagPath {
	return &TagPath{
		tag:        tag,
		tagManager: tagManager,
	}
}

func (t *TagPath) GetItems() []*types.ListItem {
	var items []*types.ListItem
	paths := t.tagManager.GetPathsForTag(t.tag)
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

func (t *TagPath) GetPath() string {
	return "tags://" + t.tag
}

func (t *TagPath) GetParentPath() string {
	return "tags://"
}

func (t *TagPath) GetName() string {
	return t.tag
}

func getDirItemCount(dirPath string) int {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0
	}
	return len(entries)
}
