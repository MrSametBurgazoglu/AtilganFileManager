package trash

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/MrSametBurgazoglu/atilgan/types"
)

const trashPath = "trash://"

type Trash struct {
	Items []TrashItem
}

func NewTrash() *Trash {
	return &Trash{
		Items: []TrashItem{},
	}
}

func (t *Trash) GetItems() []*types.ListItem {
	t.Items, _ = GetItems()
	listItems := make([]*types.ListItem, len(t.Items))
	for i, item := range t.Items {
		listItems[i] = &types.ListItem{
			Name: item.Name,
			Path: trashPath + item.Name,
		}
	}
	return listItems
}

func (t *Trash) GetPath() string {
	return trashPath
}

func (t *Trash) GetParentPath() string {
	return ""
}

func (t *Trash) GetName() string {
	return "Trash"
}

type TrashItem struct {
	Name         string
	OriginalPath string
	DeletionDate string
}

func getTrashDir() (string, error) {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdgDataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(xdgDataHome, "Trash"), nil
}

func GetItems() ([]TrashItem, error) {
	trashDir, err := getTrashDir()
	if err != nil {
		return nil, err
	}

	infoDir := filepath.Join(trashDir, "info")
	files, err := os.ReadDir(infoDir)
	if err != nil {
		return nil, err
	}

	var items []TrashItem
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".trashinfo") {
			infoPath := filepath.Join(infoDir, file.Name())
			item, err := parseTrashInfo(infoPath)
			if err != nil {
				continue
			}
			items = append(items, *item)
		}
	}

	return items, nil
}

func GetItemInfo(fileName string) (*TrashItem, error) {

	trashDir, err := getTrashDir()
	if err != nil {
		println(err.Error())
		return nil, err
	}

	infoPath := filepath.Join(trashDir, "info", fileName+".trashinfo")
	return parseTrashInfo(infoPath)
}

func parseTrashInfo(infoPath string) (*TrashItem, error) {
	file, err := os.Open(infoPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var path, deletionDate string

	if !scanner.Scan() || scanner.Text() != "[Trash Info]" {
		return nil, fmt.Errorf("invalid trashinfo file: missing header")
	}

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "Path":
			decodedPath, err := url.PathUnescape(value)
			if err != nil {
				return nil, fmt.Errorf("failed to decode path: %w", err)
			}
			path = decodedPath
		case "DeletionDate":
			deletionDate = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if path == "" {
		return nil, fmt.Errorf("invalid trashinfo file: missing Path")
	}

	baseName := strings.TrimSuffix(filepath.Base(infoPath), ".trashinfo")

	return &TrashItem{
		Name:         baseName,
		OriginalPath: path,
		DeletionDate: deletionDate,
	}, nil
}

func Restore(itemName string) error {
	trashDir, err := getTrashDir()
	if err != nil {
		return err
	}

	infoPath := filepath.Join(trashDir, "info", itemName+".trashinfo")
	item, err := parseTrashInfo(infoPath)
	if err != nil {
		return err
	}

	sourcePath := filepath.Join(trashDir, "files", itemName)

	destDir := filepath.Dir(item.OriginalPath)
	if _, err := os.Stat(destDir); os.IsNotExist(err) {
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	if err := os.Rename(sourcePath, item.OriginalPath); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	if err := os.Remove(infoPath); err != nil {
		return fmt.Errorf("failed to remove .trashinfo file: %w", err)
	}

	return nil
}
