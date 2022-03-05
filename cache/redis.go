package cache

import (
	"context"
	"log"
	"time"

	redis "github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client    redis.UniversalClient
	ttl       time.Duration
	appPrefix string
}

const apqPrefix = "apq:"

func NewRedis(redisAddress string, appPrefix string, ttl time.Duration) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddress,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		log.Panicf("could not create cache: %v", err)
	}

	return &RedisCache{
		client:    client,
		ttl:       ttl,
		appPrefix: apqPrefix + ":",
	}
}

func (c *RedisCache) Add(ctx context.Context, key string, value interface{}) {
	c.client.Set(ctx, buildKey(c.appPrefix, key), value, c.ttl)
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, bool) {
	s, err := c.client.Get(ctx, buildKey(c.appPrefix, key)).Result()
	if err != nil {
		return struct{}{}, false
	}
	return s, true
}

func buildKey(appPrefix, key string) string {
	return apqPrefix + appPrefix + key
}

func BuildRedisAddress(host, port string) string {
	return host + ":" + port
}
