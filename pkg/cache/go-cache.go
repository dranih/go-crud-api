package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type GocacheCache struct {
	prefix string
	config string
	BaseCache
	cache *gocache.Cache
}

func NewGocacheCache(prefix, config string) *GocacheCache {
	c := gocache.New(gocache.NoExpiration, 0)
	return &GocacheCache{prefix: prefix, config: config, cache: c}
}

func (gc *GocacheCache) Set(key, value string, ttl int32) bool {
	gc.cache.Set(gc.prefix+key, value, time.Duration(ttl))
	return true
}

func (gc *GocacheCache) Get(key string) string {
	if val, found := gc.cache.Get(gc.prefix + key); found {
		return val.(string)
	}
	return ""
}

func (gc *GocacheCache) Clear() bool {
	gc.cache.Flush()
	return true
}
