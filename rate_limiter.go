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
	Allow(ctx context.Context, opts *Options, key string) (Headers, int, error)
}

type rateLimiter struct {
	store Store
	mu    sync.Mutex
}

func New(store Store) RateLimiter {
	return &rateLimiter{
		store: store,
	}
}

func (r *rateLimiter) Allow(ctx context.Context, opts *Options, key string) (Headers, int, error) {
	var (
		timestamp = time.Now()
		curWindow = r.curWindow(opts, timestamp)
	)
	rate, statusCode, err := r.calculate(ctx, opts, key, timestamp, curWindow)
	return newHeaders(opts, timestamp, curWindow, rate), statusCode, err
}

func (r *rateLimiter) calculate(ctx context.Context, opts *Options, key string,
	timestamp time.Time, curWindow int64,
) (float64, int, error) {
	var (
		prevKey = r.rateKey(key, curWindow-1)
		curKey  = r.rateKey(key, curWindow)
	)

	r.mu.Lock()
	defer r.mu.Unlock()

	prev, cur, err := r.store.Get(ctx, prevKey, curKey)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	rate := r.prevWindowPercent(opts, timestamp)*float64(prev) + float64(cur)
	if int(rate) >= opts.Rate {
		return rate, http.StatusTooManyRequests, nil
	}

	err = r.store.Inc(ctx, curKey)
	if err != nil {
		return rate, http.StatusInternalServerError, err
	}

	return rate, http.StatusOK, nil
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
