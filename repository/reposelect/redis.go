package repository

import (
	"log"

	"github.com/WilfredDube/fxtract-backend/configuration"
	"github.com/WilfredDube/fxtract-backend/lib/helper"
	"github.com/go-redis/redis"
)

var cache *redis.Client
var cacheChannel chan string

func SetUpRedis(config configuration.ServiceConfig) *redis.Client {
	cache = redis.NewClient(&redis.Options{
		Addr: config.RedisHost + ":" + config.RedisPort,
		DB:   0,
	})

	setUpCacheChannel()

	return cache
}

func setUpCacheChannel() {
	cacheChannel = make(chan string)

	go func(ch chan string) {
		for {
			if err := cache.Del(<-ch).Err(); err != nil {
				log.Panic("Failed to clear cache")
			}
		}
	}(cacheChannel)
}

func ClearCache(key string) {
	helper.DPrintf("Clearing cache: %s", key)
	cacheChannel <- key
}
