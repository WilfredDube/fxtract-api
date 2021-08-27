package repository

import (
	"log"

	"github.com/go-redis/redis"
)

var cache *redis.Client
var cacheChannel chan string

func SetUpRedis() *redis.Client {
	cache = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
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
			} else {
				log.Println("Cache cleared")
			}
		}
	}(cacheChannel)
}

func ClearCache(key string) {
	log.Println("Clearing cache.............")
	cacheChannel <- key
}
