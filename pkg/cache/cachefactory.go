package cache

func Create(cacheType, prefix, config string) Cache {
	var cache interface{ Cache }
	switch cacheType {
	//Keeping tempfile for compatibility
	case "TempFile", "Gocache":
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
