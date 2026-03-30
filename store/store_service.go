package store

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// Define the struct wrapper around raw Redis client
type StorageService struct {
	redisClient *redis.Client
}

// Top level declarations for the storeService and Redis context
var (
	storeService = &StorageService{}
	ctx = context.Background()
)

// Note that in a real world usage, the cache duration should not have an expiration time,
// an LRU policy config should be set where the values that are retrieved less often are
// purged automatically from the cache and stored back in RDBMS whenever the cache is full.
const cacheDuration = 6 * time.Hour

func getRedisAddress() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}

	if host := os.Getenv("REDIS_HOST"); host != "" {
		fmt.Println("Using REDIS_HOST environment variable: ", host)
		return host + ":6379"
	}

	return "localhost:6379"
}

func getRedisDB() int {
	raw := os.Getenv("REDIS_DB")

	if raw == "" {
		return 0
	}

	n, err := strconv.Atoi(raw)

	if err != nil {
		return 0
	}

	return n
}

// Initialize the store service and return a store pointer
func InitializeStore() *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr: getRedisAddress(),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: getRedisDB(),					// getRedisDB() returns the Redis DB number to use
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis: %v", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	return storeService
}

/* 
We want to be able to save the mapping between the originalUrl
and the generated shortURL
*/
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, cacheDuration).Err()

	if err != nil {
		panic(fmt.Sprintf("Failed saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
}

/*
We should be able to retrieve the initial long URL once the short is provided.
This is when the users will be calling the shortlink in the url, so what we 
need to do here is to retrieve the long url and think about redirect.
*/
func RetrieveInitialUrl(shortUrl string) string {
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()

	if err != nil {
		panic(fmt.Sprintf("Failed RetrieveInitialUrl url | Error %v - shortUrl: %s\n", err, shortUrl))
	}

	return result
}
