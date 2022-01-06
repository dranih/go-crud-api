package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	prefix string
	config string
	BaseCache
	redisClient *redis.Client
	ctx         context.Context
}

func NewRedisCache(prefix, config string) *RedisCache {
	var redisConfig *redis.Options
	if err := json.Unmarshal([]byte(config), &redisConfig); err != nil {
		log.Printf("Error loading Redis configuration : %v", err)
		return nil
	} else {
		rdb := redis.NewClient(redisConfig)
		return &RedisCache{prefix: prefix, config: config, redisClient: rdb, ctx: context.Background()}
	}
}

func (rc *RedisCache) Set(key, value string, ttl int32) bool {
	if err := rc.redisClient.Set(rc.ctx, rc.prefix+key, value, time.Duration(ttl)).Err(); err != nil {
		log.Printf("Caching error : %v", err)
		return false
	}
	return true
}

func (rc *RedisCache) Get(key string) string {
	if item, err := rc.redisClient.Get(rc.ctx, rc.prefix+key).Result(); err != nil {
		log.Printf("Caching error : %v", err)
		return ""
	} else {
		return item
	}
}

func (rc *RedisCache) Clear() bool {
	if err := rc.redisClient.FlushDB(rc.ctx).Err(); err != nil {
		return false
	}
	return true
}
