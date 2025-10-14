package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(host string, port int, password string) *RedisClient {
	addr := fmt.Sprintf("%s:%d", host, port)
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisClient{client: rdb}
}

func (r *RedisClient) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	return r.client.Set(ctx, "blacklist:"+token, "1", ttl).Err()
}

func (r *RedisClient) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	_, err := r.client.Get(ctx, "blacklist:"+token).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
