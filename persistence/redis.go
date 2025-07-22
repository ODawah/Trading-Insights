package persistence

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func ConnectRedis() error {
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
		return err
	}
	return nil
}
