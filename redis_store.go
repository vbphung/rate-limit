package ratelimit

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(client *redis.Client) Store {
	return &RedisStore{client}
}

func (r *RedisStore) Get(ctx context.Context, prevKey, curKey string) (int, int, error) {
	prev, err := r.get(ctx, prevKey)
	if err != nil {
		return 0, 0, err
	}

	cur, err := r.get(ctx, curKey)
	if err != nil {
		return 0, 0, err
	}

	return prev, cur, nil
}

func (r *RedisStore) get(ctx context.Context, key string) (int, error) {
	if v, err := r.client.Get(ctx, key).Result(); err == nil {
		return strconv.Atoi(v)
	} else if err == redis.Nil {
		if err = r.client.Set(ctx, key, 0, time.Second).Err(); err != nil {
			return 0, err
		}

		return 0, nil
	} else {
		return 0, err
	}
}

func (r *RedisStore) Inc(ctx context.Context, key string) error {
	return r.client.Incr(ctx, key).Err()
}
