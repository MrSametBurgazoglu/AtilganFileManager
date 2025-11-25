package fileops

import (
	"time"
)

func GetModifiedTimeAsString(modifiedTime time.Time) string {
	return modifiedTime.Format("2006-01-02 15:04:05")
}
