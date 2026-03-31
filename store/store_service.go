package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Define the struct wrapper around raw Redis client
type StorageService struct {
	redisClient *redis.Client
}

var db *sql.DB

// Top level declarations for the storeService and Redis context
var (
	storeService = &StorageService{}
	ctx = context.Background()
)

func InitializeDB() {
	dsn := os.Getenv("DATABASE_URL")
	
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	fmt.Println("database connection pool established")
}

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
	
	// 1. Write to Postgre (Superbase)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, 
		`insert into short_urls (short_code, long_url, user_id)
		values ($1, $2, $3)
		on conflict (short_code) do update set long_url = excluded.long_url`,
	shortUrl, originalUrl, userId,
	)
	if err != nil {
		panic(fmt.Sprintf("failed to save URL mapping in Postgres: %v", err))
	}

	// 2. Cache in Redis
	if err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, cacheDuration).Err(); err != nil {
		// Here we can log instead of panic if you want DB to be authoritative
		panic(fmt.Sprintf("failed caching in redis and saving key url | Error: %v - shortUrl: %s - originalUrl: %s\n", err, shortUrl, originalUrl))
	}
}

/*
We should be able to retrieve the initial long URL once the short is provided.
This is when the users will be calling the shortlink in the url, so what we 
need to do here is to retrieve the long url and think about redirect.
*/
func RetrieveInitialUrl(shortUrl string) (string, error) {

	// 1. Check if the short URL is cached in Redis
	result, err := storeService.redisClient.Get(ctx, shortUrl).Result()
	if err == nil {
		return result, nil
	}
	if err != nil && err != redis.Nil {
		// true redis error
		return "", fmt.Errorf("failed to get url from redis | redis error: %v", err)
	}

	// 2. Cache miss: go to Postgres
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var longURL string
	qerr := db.QueryRowContext(ctx, `select long_url from short_urls where short_code = $1`, shortUrl).Scan(&longURL)

	if qerr == sql.ErrNoRows {
		// not found anywhere: let the handler turn this into a 404
		return "", fmt.Errorf("short URL not found in database | shortUrl: %s", shortUrl)
	}
	if qerr != nil {
		return "", fmt.Errorf("failed to get url from postgres | postgres error: %v; shortUrl: %s", qerr, shortUrl)
	}

	// 3. Fill cache again 
	_ = storeService.redisClient.Set(ctx, shortUrl, longURL, cacheDuration).Err()

	return longURL, nil // success
}
