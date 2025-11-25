package cache

import (
	"sync"

	"github.com/diamondburned/gotk4/pkg/gdk/v4"
)

var (
	imageCache = make(map[string][]*gdk.Texture)
	mutex      = &sync.RWMutex{}
)

func Add(key string, textures []*gdk.Texture) {
	mutex.Lock()
	defer mutex.Unlock()
	imageCache[key] = textures
}

func Get(key string) ([]*gdk.Texture, bool) {
	mutex.RLock()
	defer mutex.RUnlock()
	textures, found := imageCache[key]
	return textures, found
}

func Clear() {
	mutex.Lock()
	defer mutex.Unlock()
	imageCache = make(map[string][]*gdk.Texture)
}
