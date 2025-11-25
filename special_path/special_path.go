package special_path

import (
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/recent"
	"github.com/MrSametBurgazoglu/atilgan/tag"
	"github.com/MrSametBurgazoglu/atilgan/trash"
)

type SpecialPathManager struct {
	Paths         map[string]IPath
	tagManager    *tag.TagManager
	recentManager *recent.RecentManager
}

func NewSpecialPathManager() (*SpecialPathManager, error) {
	tagManager, err := tag.NewTagManager()
	if err != nil {
		println(err.Error())
		return nil, err
	}
	recentManager, err := recent.NewRecentManager()
	if err != nil {
		return nil, err
	}
	return &SpecialPathManager{
		Paths: map[string]IPath{
			"trash":  trash.NewTrash(),
			"tags":   tag.NewTagsPath(tagManager),
			"recent": recent.NewRecentPath(recentManager),
		},
		tagManager:    tagManager,
		recentManager: recentManager,
	}, nil
}

func (spm *SpecialPathManager) GetPath(path string) IPath {
	if strings.HasPrefix(path, "tags://") {
		parts := strings.Split(strings.TrimPrefix(path, "tags://"), "/")
		if len(parts) == 1 && parts[0] != "" {
			return tag.NewTagPath(parts[0], spm.tagManager)
		}
		return spm.Paths["tags"]
	}
	if strings.HasPrefix(path, "trash://") {
		return spm.Paths["trash"]
	}
	if strings.HasPrefix(path, "recent://") {
		return spm.Paths["recent"]
	}
	return nil
}

func (spm *SpecialPathManager) AddRecentPath(path string) {
	spm.recentManager.AddPath(path)
}

func (spm *SpecialPathManager) GetTagManager() *tag.TagManager {
	return spm.tagManager
}
