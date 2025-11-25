package recent

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RecentManager struct {
	Paths  []string
	dbPath string
}

func NewRecentManager() (*RecentManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dbPath := filepath.Join(configDir, "atilgan", "recent.json")

	rm := &RecentManager{
		dbPath: dbPath,
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return nil, err
		}
		if err := rm.save(); err != nil {
			return nil, err
		}
	} else {
		if err := rm.load(); err != nil {
			return nil, err
		}
	}

	return rm, nil
}

func (rm *RecentManager) load() error {
	data, err := os.ReadFile(rm.dbPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &rm.Paths)
}

func (rm *RecentManager) save() error {
	data, err := json.MarshalIndent(rm.Paths, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rm.dbPath, data, 0644)
}

func (rm *RecentManager) AddPath(path string) {
	for i, p := range rm.Paths {
		if p == path {
			rm.Paths = append(rm.Paths[:i], rm.Paths[i+1:]...)
			break
		}
	}

	rm.Paths = append([]string{path}, rm.Paths...)

	if len(rm.Paths) > 100 {
		rm.Paths = rm.Paths[:100]
	}

	rm.save()
}

func (rm *RecentManager) GetPaths() []string {
	return rm.Paths
}
