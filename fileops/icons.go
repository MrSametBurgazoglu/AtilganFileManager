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

func IsImage(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".png") ||
		strings.HasSuffix(fileName, ".jpg") ||
		strings.HasSuffix(fileName, ".jpeg") ||
		strings.HasSuffix(fileName, ".gif") ||
		strings.HasSuffix(fileName, ".webp") ||
		strings.HasSuffix(fileName, ".svg")
}

func IsText(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".txt") ||
		strings.HasSuffix(fileName, ".mod") ||
		strings.HasSuffix(fileName, ".sum")
}

func IsVideo(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".mp4") ||
		strings.HasSuffix(fileName, ".mkv") ||
		strings.HasSuffix(fileName, ".mov") ||
		strings.HasSuffix(fileName, ".avi") ||
		strings.HasSuffix(fileName, ".webm")
}

func IsMusic(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".mp3") ||
		strings.HasSuffix(fileName, ".ogg") ||
		strings.HasSuffix(fileName, ".wav") ||
		strings.HasSuffix(fileName, ".flac") ||
		strings.HasSuffix(fileName, ".m4a")
}

func IsMedia(fileName string) bool {
	return IsVideo(fileName) || IsMusic(fileName)
}

func GetVideosPath() string {
	return videos
}

func GetPicturesPath() string {
	return pictures
}

func GetMusicPath() string {
	return music
}

func IsDocument(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".pdf") ||
		strings.HasSuffix(fileName, ".epub") ||
		strings.HasSuffix(fileName, ".mobi") ||
		strings.HasSuffix(fileName, ".docx") ||
		strings.HasSuffix(fileName, ".xlsx") ||
		strings.HasSuffix(fileName, ".pptx")
}

func IsCode(fileName string) bool {
	fileName = strings.ToLower(fileName)
	return strings.HasSuffix(fileName, ".go") ||
		strings.HasSuffix(fileName, ".json") ||
		strings.HasSuffix(fileName, ".yaml") ||
		strings.HasSuffix(fileName, ".yml") ||
		strings.HasSuffix(fileName, ".env") ||
		strings.HasSuffix(fileName, "dockerfile") ||
		strings.HasSuffix(fileName, ".js") ||
		strings.HasSuffix(fileName, ".ts") ||
		strings.HasSuffix(fileName, ".py") ||
		strings.HasSuffix(fileName, ".java") ||
		strings.HasSuffix(fileName, ".c") ||
		strings.HasSuffix(fileName, ".cpp") ||
		strings.HasSuffix(fileName, ".h") ||
		strings.HasSuffix(fileName, ".hpp") ||
		strings.HasSuffix(fileName, ".rs") ||
		strings.HasSuffix(fileName, ".rb") ||
		strings.HasSuffix(fileName, ".php") ||
		strings.HasSuffix(fileName, ".swift") ||
		strings.HasSuffix(fileName, ".kt") ||
		strings.HasSuffix(fileName, ".kts") ||
		strings.HasSuffix(fileName, ".sh") ||
		strings.HasSuffix(fileName, ".bat") ||
		strings.HasSuffix(fileName, ".md")
}
