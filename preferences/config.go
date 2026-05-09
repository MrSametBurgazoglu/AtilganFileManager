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

const (
	DefaultTheme             = "system"
	DefaultShowHidden        = false
	DefaultIconSize          = 48
	DefaultSidebarWidth      = 180
	DefaultViewMode          = "grid"
	DefaultConfirmDelete     = true
	DefaultEnablePreviewPane = true
	DefaultTerminalEmulator  = "gnome-terminal"
	DefaultEnableThumbnails  = true
	DefaultSortOrder         = "name"
)

func DefaultConfig() *Config {
	return &Config{
		Theme:             DefaultTheme,
		ShowHidden:        DefaultShowHidden,
		IconSize:          DefaultIconSize,
		SidebarWidth:      DefaultSidebarWidth,
		DefaultViewMode:   DefaultViewMode,
		ConfirmDelete:     DefaultConfirmDelete,
		EnablePreviewPane: DefaultEnablePreviewPane,
		TerminalEmulator:  DefaultTerminalEmulator,
		EnableThumbnails:  DefaultEnableThumbnails,
		DefaultSortOrder:  DefaultSortOrder,
	}
}

func GetConfigPath() string {
	return filepath.Join(xdg.ConfigHome, "atilgan", "config.json")
}

func LoadConfig() *Config {
	config := DefaultConfig()
	path := GetConfigPath()
	
	data, err := os.ReadFile(path)
	if err != nil {
		return config
	}

	if err := json.Unmarshal(data, config); err != nil {
		// If config is corrupted, return default
		return DefaultConfig()
	}
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
