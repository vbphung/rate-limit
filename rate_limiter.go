package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Options struct {
	Rate   int
	Window time.Duration
	Epoch  time.Time
}

type RateLimiter interface {
	Allow(ctx context.Context, opts *Options, key string) (int, error)
}

type rateLimiter struct {
	store Store
	mx    sync.Mutex
}

func New(store Store) RateLimiter {
	return &rateLimiter{
		store: store,
	}
}

func (r *rateLimiter) Allow(ctx context.Context, opts *Options, key string) (int, error) {
	var (
		timestamp = time.Now()
		curWindow = r.curWindow(opts, timestamp)
		prevKey   = r.rateKey(key, curWindow-1)
		curKey    = r.rateKey(key, curWindow)
	)

	r.mx.Lock()
	defer r.mx.Unlock()

	prev, cur, err := r.store.Get(ctx, prevKey, curKey)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	rate := r.prevWindowPercent(opts, timestamp)*float64(prev) + float64(cur)
	if int(rate) >= opts.Rate {
		return http.StatusTooManyRequests, nil
	}

	err = r.store.Inc(ctx, curKey)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (r *rateLimiter) curWindow(opts *Options, t time.Time) int64 {
	diff := t.Sub(opts.Epoch)
	return int64(diff / opts.Window)
}

func (r *rateLimiter) prevWindowPercent(opts *Options, t time.Time) float64 {
	elapsed := t.Sub(opts.Epoch) % opts.Window
	return 1 - float64(elapsed)/float64(opts.Window)
}

func (r *rateLimiter) rateKey(key string, window int64) string {
	return fmt.Sprintf("@rate:%s:%d", key, window)
}
