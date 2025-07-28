package persistence

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client *redis.Client
}

type RedisRepository interface {
	Set(ctx context.Context, key string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
}

func ConnectRedis() (RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // use default DB
		Protocol: 2,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Could not connect to Redis: %v", err))
		return nil, err
	}

	return &redisRepository{client: client}, nil
}

func (r *redisRepository) Set(ctx context.Context, key string, value interface{}) error {
	err := r.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("error setting value in Redis: %w", err)
	}
	return nil
}

func (r *redisRepository) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("error getting value from Redis: %w", err)
	}
	return value, nil
}
