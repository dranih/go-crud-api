package cache

import (
	"fmt"
	"log"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

type MemcacheCache struct {
	prefix string
	config string
	BaseCache
	memcache *memcache.Client
}

func NewMemcacheCache(prefix, config string) *MemcacheCache {
	var address string
	if config == "" {
		address = "localhost:11211"
	} else if strings.Index(config, ":") <= 0 {
		address = fmt.Sprintf("%s:11211", config)
	} else {
		address = config

	}
	mc := memcache.New(address)
	return &MemcacheCache{prefix: prefix, config: config, memcache: mc}
}

func (mc *MemcacheCache) Set(key, value string, ttl int32) bool {
	if err := mc.memcache.Set(&memcache.Item{Key: mc.prefix + key, Value: []byte(value), Expiration: ttl}); err != nil {
		log.Printf("Caching error : %v", err)
		return false
	}
	return true
}

func (mc *MemcacheCache) Get(key string) string {
	if item, err := mc.memcache.Get(mc.prefix + key); err != nil {
		log.Printf("Caching error : %v", err)
		return ""
	} else {
		return string(item.Value)
	}
}

func (mc *MemcacheCache) Clear() bool {
	if err := mc.memcache.FlushAll(); err != nil {
		return false
	}
	return true
}
