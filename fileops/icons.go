package fileops

import (
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

var fileTypeIcons = map[string]string{
	".go":   "text-x-go",
	".py":   "text-x-python",
	".js":   "application-javascript",
	".ts":   "application-typescript",
	".json": "application-json",
	".md":   "text-markdown",
	".txt":  "text-plain",
	".pdf":  "application-pdf",
	".png":  "image-x-generic",
	".jpg":  "image-x-generic",
	".jpeg": "image-jpeg",
	".gif":  "image-gif",
	".svg":  "image-svg+xml",
	".zip":  "application-zip",
	".gz":   "application-gzip",
	".tar":  "application-x-tar",
	".rar":  "application-x-rar",
	".mp3":  "audio-mpeg",
	".ogg":  "audio-ogg",
	".wav":  "audio-x-wav",
	".mp4":  "video-x-generic",
	".mkv":  "video-x-matroska",
	".mov":  "video-quicktime",
	".avi":  "video-x-msvideo",
}

var (
	home        string = xdg.Home
	desktop     string = xdg.UserDirs.Desktop
	downloads   string = xdg.UserDirs.Download
	documents   string = xdg.UserDirs.Documents
	pictures    string = xdg.UserDirs.Pictures
	music       string = xdg.UserDirs.Music
	videos      string = xdg.UserDirs.Videos
	publicShare string = xdg.UserDirs.PublicShare
	templates   string = xdg.UserDirs.Templates
)

var folderPathIcons = map[string]string{
	"/":         "folder-root",
	home:        "user-home",
	"trash://":  "user-trash",
	"recent://": "document-open-recent",
	"tags://":   "tag",
	desktop:     "user-desktop",
	documents:   "folder-documents",
	downloads:   "folder-download",
	music:       "folder-music",
	pictures:    "folder-pictures",
	videos:      "folder-videos",
	publicShare: "folder-publicshare",
	templates:   "folder-templates",
}

func GetIconForFile(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if icon, ok := fileTypeIcons[ext]; ok {
		return icon
	}
	return "text-x-generic" // Default icon
}

func GetIconForFolderSymbolic(folderPath string) string {
	iconName := GetIconForFolder(folderPath)
	return iconName + "-symbolic"
}

func GetIconForFolder(folderPath string) string {
	if icon, ok := folderPathIcons[folderPath]; ok {
		return icon
	}
	return "folder" // Default icon
}
