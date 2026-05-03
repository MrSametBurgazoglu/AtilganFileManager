package fileops

import "fmt"

func GetFileSizeAsString(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d bytes", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.0f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.0f MB", float64(size)/(1024*1024))
	} else {
		return fmt.Sprintf("%.0f GB", float64(size)/(1024*1024*1024))
	}
}
