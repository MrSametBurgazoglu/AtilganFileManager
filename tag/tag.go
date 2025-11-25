package tag

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
)

type TagManager struct {
	Tags   map[string][]string
	dbPath string
}

func NewTagManager() (*TagManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(configDir, "atilgan", "tags.json")

	tm := &TagManager{
		dbPath: dbPath,
		Tags:   make(map[string][]string),
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return nil, err
		}
		if err := tm.save(); err != nil {
			return nil, err
		}
	} else {
		if err := tm.load(); err != nil {
			return nil, err
		}
	}
	return tm, nil
}

func (tm *TagManager) load() error {
	data, err := os.ReadFile(tm.dbPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &tm.Tags)
}

func (tm *TagManager) save() error {
	data, err := json.MarshalIndent(tm.Tags, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tm.dbPath, data, 0644)
}

func (tm *TagManager) AddTag(path string, tag string) {
	tags, ok := tm.Tags[path]
	if !ok {
		tags = []string{}
	}
	if slices.Contains(tags, tag) {
		return
	}
	tm.Tags[path] = append(tags, tag)
	tm.save()
}

func (tm *TagManager) RemoveTag(path string, tag string) {
	tags, ok := tm.Tags[path]
	if !ok {
		return
	}
	newTags := []string{}
	for _, t := range tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}
	tm.Tags[path] = newTags
	tm.save()
}

func (tm *TagManager) GetTags(path string) []string {
	return tm.Tags[path]
}

func (tm *TagManager) GetPathsForTag(tag string) []string {
	var paths []string
	for path, tags := range tm.Tags {
		for _, t := range tags {
			if t == tag {
				paths = append(paths, path)
				break
			}
		}
	}
	sort.Strings(paths)
	return paths
}

func (tm *TagManager) GetAllTags() []string {
	tagSet := make(map[string]struct{})
	for _, tags := range tm.Tags {
		for _, tag := range tags {
			tagSet[tag] = struct{}{}
		}
	}
	var allTags []string
	for tag := range tagSet {
		allTags = append(allTags, tag)
	}
	sort.Strings(allTags)
	return allTags
}
