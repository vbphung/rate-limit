package ratelimit

import "context"

type Store interface {
	Get(ctx context.Context, prevKey, curKey string) (int, int, error)
	Inc(ctx context.Context, key string) error
}
