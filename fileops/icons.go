package fileops

import (
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

var fileTypeIcons = map[string]string{
	".go":   "text-x-script",
	".py":   "text-x-script",
	".js":   "text-x-script",
	".ts":   "text-x-script",
	".json": "text-x-generic",
	".md":   "text-x-generic",
	".txt":  "text-x-generic",
	".pdf":  "text-x-generic",
	".png":  "image-x-generic",
	".jpg":  "image-x-generic",
	".jpeg": "image-x-generic",
	".gif":  "image-x-generic",
	".svg":  "image-x-generic",
	".zip":  "text-x-generic",
	".gz":   "text-x-generic",
	".tar":  "text-x-generic",
	".rar":  "text-x-generic",
	".mp3":  "text-x-generic",
	".ogg":  "text-x-generic",
	".wav":  "text-x-generic",
	".mp4":  "image-x-generic",
	".mkv":  "image-x-generic",
	".mov":  "image-x-generic",
	".avi":  "image-x-generic",
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

var fileDescriptions = map[string]string{
	".go":   "Go Source File",
	".py":   "Python Script",
	".js":   "JavaScript File",
	".ts":   "TypeScript File",
	".json": "JSON Data",
	".md":   "Markdown Document",
	".txt":  "Plain Text File",
	".pdf":  "PDF Document",
	".png":  "PNG Image",
	".jpg":  "JPEG Image",
	".jpeg": "JPEG Image",
	".gif":  "GIF Image",
	".svg":  "SVG Image",
	".webp": "WebP Image",
	".zip":  "ZIP Archive",
	".gz":   "Gzip Archive",
	".tar":  "Tar Archive",
	".rar":  "RAR Archive",
	".mp3":  "MP3 Audio",
	".ogg":  "Ogg Audio",
	".wav":  "WAV Audio",
	".mp4":  "MP4 Video",
	".mkv":  "Matroska Video",
	".mov":  "QuickTime Video",
	".avi":  "AVI Video",
	".html": "HTML Document",
	".css":  "CSS Stylesheet",
	".sh":   "Shell Script",
}

func GetFileDescription(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	if desc, ok := fileDescriptions[ext]; ok {
		return desc
	}
	if ext == "" {
		return "Binary/Executable"
	}
	return strings.ToUpper(strings.TrimPrefix(ext, ".")) + " File"
}
