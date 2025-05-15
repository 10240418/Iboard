package databases

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var REDIS_CONN *redis.Client

func InitRedis() error {
	var err error
	maxRetries := 5
	retryDelay := time.Second * 3

	for i := 0; i < maxRetries; i++ {
		err = initRedisConnection()
		if err == nil {
			log.Printf("Successfully connected to Redis on attempt %d", i+1)
			return nil
		}

		log.Printf("Failed to connect to Redis (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("failed to connect to Redis after %d attempts: %v", maxRetries, err)
}

func initRedisConnection() error {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "redis" // 默认使用容器名
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisDB := os.Getenv("REDIS_DB")
	if redisDB == "" {
		redisDB = "0"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	db, err := strconv.Atoi(redisDB)
	if err != nil {
		return fmt.Errorf("invalid redis db number: %v", err)
	}

	log.Printf("Attempting to connect to Redis at %s:%s", redisHost, redisPort)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:     redisPassword,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %v", err)
	}

	REDIS_CONN = rdb
	return nil
}

func CloseRedis() error {
	if REDIS_CONN != nil {
		log.Println("Closing Redis connection...")
		return REDIS_CONN.Close()
	}
	return nil
}
