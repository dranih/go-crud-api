package cache

func Create(cacheType, prefix, config string) Cache {
	var cache interface{ Cache }
	switch cacheType {
	case "TempFile":
		cache = NewGocacheCache(prefix, config)
	case "Redis":
		cache = NewRedisCache(prefix, config)
	case "Memcache", "Memcached":
		cache = NewMemcacheCache(prefix, config)
	default:
		cache = &NoCache{}
	}
	return cache
}
