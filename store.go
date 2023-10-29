package cider

import (
	"context"
	"errors"
	"sync"
)

type Storer interface {
	/// Gets value by key.
	Get(ctx context.Context, key string) (result []byte, err error)
	// Sets value by key.
	Set(ctx context.Context, key string, value []byte) (err error)
	// Deletes keys. Returns the number of deleted keys.
	Del(ctx context.Context, keys []string) (deleted int, err error)
	// Checks if keys exist in database. Returns the number of keys found.
	Exists(ctx context.Context, keys []string) (found int, err error)
}

type store struct {
	mu *sync.RWMutex
	db map[string][]byte
}

func NewStore() *store {
	return &store{
		mu: &sync.RWMutex{},
		db: make(map[string][]byte),
	}
}

func (s *store) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.db[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	return val, nil
}

func (s *store) Set(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = value

	return nil
}

func (s *store) Del(ctx context.Context, keys []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deletes := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			delete(s.db, key)
			deletes++
		}
	}
	return deletes, nil
}

func (s *store) Exists(ctx context.Context, keys []string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	found := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			found++
		}
	}
	return found, nil
}
