package cache

import "time"

type Cache interface {
	Set(string, string, int32) bool
	Get(string) string
	Clear() bool
	Ping() int
}

type BaseCache struct{}

func (bc *BaseCache) Set(key, value string, ttl int32) bool {
	return true
}

func (bc *BaseCache) Get(key string) string {
	return ""
}

func (bc *BaseCache) Clear() bool {
	return true
}

func (bc *BaseCache) Ping() int {
	start := time.Now()
	bc.Get("__ping__")
	t := time.Now()
	elapsed := t.Sub(start)
	return int(elapsed.Milliseconds())
}

type NoCache struct {
	BaseCache
}
