package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MrSametBurgazoglu/atilgan/types"
)

func GetGroupForTime(modTime time.Time) string {
	now := time.Now()
	duration := now.Sub(modTime)

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if modTime.After(todayStart) {
		return "Today"
	}

	if duration.Hours() <= 24 {
		return "Last 24 hours"
	}

	if duration.Hours() <= 24*7 {
		return "Last Week"
	}

	if duration.Hours() <= 24*30 {
		return "Last Month"
	}

	return "Later"
}

func CalculateSizeThresholds(entries []os.DirEntry) []int64 {
	var sizes []int64
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				sizes = append(sizes, info.Size())
			}
		}
	}
	if len(sizes) == 0 {
		return nil
	}

	sort.Slice(sizes, func(i, j int) bool { return sizes[i] < sizes[j] })

	candidateThresholds := []int64{
		1024, 2 * 1024, 5 * 1024,
		10 * 1024, 20 * 1024, 50 * 1024,
		100 * 1024, 200 * 1024, 500 * 1024,
		1024 * 1024, 2 * 1024 * 1024, 5 * 1024 * 1024,
		10 * 1024 * 1024, 20 * 1024 * 1024, 50 * 1024 * 1024,
		100 * 1024 * 1024, 200 * 1024 * 1024, 500 * 1024 * 1024,
		1024 * 1024 * 1024, 2 * 1024 * 1024 * 1024, 5 * 1024 * 1024 * 1024,
	}

	usedThresholds := make([]int64, 0)
	lastThreshold := int64(-1)
	for _, size := range sizes {
		for _, t := range candidateThresholds {
			if size < t {
				if t != lastThreshold {
					usedThresholds = append(usedThresholds, t)
					lastThreshold = t
				}
				break
			}
		}
		if size >= candidateThresholds[len(candidateThresholds)-1] {
			t := candidateThresholds[len(candidateThresholds)-1] * 10 // Just some large value
			if t != lastThreshold {
				usedThresholds = append(usedThresholds, t)
				lastThreshold = t
			}
		}
	}

	// Simplify thresholds if too many to ensure max 7 categories
	for len(usedThresholds) > 6 {
		newThresholds := make([]int64, 0)
		for i := 0; i < len(usedThresholds); i++ {
			if i%2 == 0 || i == len(usedThresholds)-1 {
				newThresholds = append(newThresholds, usedThresholds[i])
			}
		}
		usedThresholds = newThresholds
	}

	return usedThresholds
}

func GetGroupForSize(size int64, thresholds []int64) string {
	if len(thresholds) == 0 {
		return "Files"
	}
	for i, t := range thresholds {
		if size < t {
			if i == 0 {
				return fmt.Sprintf("< %s", GetFileSizeAsString(t))
			}
			return fmt.Sprintf("%s - %s", GetFileSizeAsString(thresholds[i-1]), GetFileSizeAsString(t))
		}
	}
	return fmt.Sprintf("> %s", GetFileSizeAsString(thresholds[len(thresholds)-1]))
}

func GetGroupForType(entry os.DirEntry) string {
	if entry.IsDir() {
		return "Directories"
	}
	fileType := GetFileType(entry)
	switch fileType {
	case types.TypeExec:
		return "Executables"
	case types.TypeDoc:
		return "Documents"
	case types.TypeMedia:
		return "Media"
	default:
		return "Extra"
	}
}

func GetFileType(entry os.DirEntry) types.FileType {
	fileName := entry.Name()
	if strings.HasPrefix(fileName, ".") {
		return types.TypeHidden
	}

	if strings.HasPrefix(fileName, "~") {
		return types.TypeTemp
	}

	if entry.IsDir() {
		return types.TypeDir
	}

	info, err := entry.Info()
	if err != nil {
		return types.TypeOther
	}

	if info.Mode()&0111 != 0 {
		return types.TypeExec
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".doc", ".docx", ".pdf", ".txt", ".md", ".odt", ".rtf", ".xls", ".xlsx", ".ppt", ".pptx", ".csv":
		return types.TypeDoc
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".mp3", ".mp4", ".avi", ".mkv", ".wav", ".ogg", ".flac", ".mov", ".wmv":
		return types.TypeMedia
	}

	return types.TypeOther
}
