package cache

import (
	"context"
	"time"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

type Redis struct {
	client *redis.Client
	retry  retry.Strategy
}

func NewRedis(client *redis.Client, retry retry.Strategy) *Redis {
	return &Redis{
		client: client,
		retry:  retry,
	}
}

func (r *Redis) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.client.SetWithExpiration(ctx, key, value, expiration)
}

func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	value, err := r.client.GetWithRetry(ctx, r.retry, key)
	return value, err
}
