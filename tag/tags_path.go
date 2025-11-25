package tag

import "github.com/MrSametBurgazoglu/atilgan/types"

type TagsPath struct {
	tagManager *TagManager
}

func NewTagsPath(tagManager *TagManager) *TagsPath {
	return &TagsPath{
		tagManager: tagManager,
	}
}

func (t *TagsPath) GetItems() []*types.ListItem {
	var items []*types.ListItem
	tags := t.tagManager.GetAllTags()
	for _, tag := range tags {
		items = append(items, &types.ListItem{
			Name:      tag,
			IsDir:     true,
			Path:      "tags://" + tag,
			ItemCount: len(t.tagManager.GetPathsForTag(tag)),
		})
	}
	return items
}

func (t *TagsPath) GetPath() string {
	return "tags://"
}

func (t *TagsPath) GetParentPath() string {
	return ""
}

func (t *TagsPath) GetName() string {
	return "Tags"
}
