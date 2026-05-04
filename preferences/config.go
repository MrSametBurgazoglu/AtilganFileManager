package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

type Config struct {
	Theme             string `json:"theme"`
	ShowHidden        bool   `json:"show_hidden"`
	IconSize          int    `json:"icon_size"`
	SidebarWidth      int    `json:"sidebar_width"`
	DefaultViewMode   string `json:"default_view_mode"` // "grid" or "list"
	ConfirmDelete     bool   `json:"confirm_delete"`
	EnablePreviewPane bool   `json:"enable_preview_pane"`
	TerminalEmulator  string `json:"terminal_emulator"`
	EnableThumbnails  bool   `json:"enable_thumbnails"`
	DefaultSortOrder  string `json:"default_sort_order"` // "name", "size", "time", "type"
}

func DefaultConfig() *Config {
	return &Config{
		Theme:             "system",
		ShowHidden:        false,
		IconSize:          48,
		SidebarWidth:      180,
		DefaultViewMode:   "grid",
		ConfirmDelete:     true,
		EnablePreviewPane: true,
		TerminalEmulator:  "gnome-terminal",
		EnableThumbnails:  true,
		DefaultSortOrder:  "name",
	}
}

func GetConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "atilgan", "config.json")
}

func LoadConfig() *Config {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	config := DefaultConfig()
	json.Unmarshal(data, config)
	return config
}

func (c *Config) Save() error {
	path := GetConfigPath()
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
