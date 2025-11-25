package cache

import (
	"sync"
)

type FileInfoCache struct {
	cache map[string]*FileInfo
	mutex *sync.RWMutex
}

type FileInfo struct {
	Width  int
	Height int
	Type   string
}

func NewFileInfoCache() *FileInfoCache {
	return &FileInfoCache{
		cache: make(map[string]*FileInfo),
		mutex: &sync.RWMutex{},
	}
}

func (fic *FileInfoCache) Get(path string) (*FileInfo, bool) {
	fic.mutex.RLock()
	defer fic.mutex.RUnlock()
	info, found := fic.cache[path]
	return info, found
}

func (fic *FileInfoCache) Set(path string, info *FileInfo) {
	fic.mutex.Lock()
	defer fic.mutex.Unlock()
	fic.cache[path] = info
}
