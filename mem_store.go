package ratelimit

import (
	"context"
	"sync"
)

type MemStore struct {
	rates map[string]int
	mu    sync.Mutex
}

func NewMemStore() Store {
	return &MemStore{
		rates: make(map[string]int),
	}
}

func (m *MemStore) Get(ctx context.Context, prevKey, curKey string) (int, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.rates[prevKey]; !ok {
		m.rates[prevKey] = 0
	}

	if _, ok := m.rates[curKey]; !ok {
		m.rates[curKey] = 0
	}

	return m.rates[prevKey], m.rates[curKey], nil
}

func (m *MemStore) Inc(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.rates[key]; !ok {
		m.rates[key] = 0
	}

	m.rates[key]++

	return nil
}
