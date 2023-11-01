package cider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type Storer interface {
	// Gets value by key.
	Get(ctx context.Context, key string) (value []byte, ttl int64, err error)
	// Sets value by key.
	Set(ctx context.Context, key string, value []byte, ttl int64) (err error)
	// Deletes keys. Returns the number of deleted keys.
	Del(ctx context.Context, keys []string) (deleted int, err error)
	// Checks if keys exist in database. Returns the number of keys found.
	Exists(ctx context.Context, keys []string) (found int, err error)
	// Expires a key after n seconds.
	Expire(ctx context.Context, key string, ttl int64) (result int, err error)
	// Increments a key.
	Incr(ctx context.Context, key string) (err error)
	// Decrements a key.
	Decr(ctx context.Context, key string) (err error)
}

type store struct {
	mu *sync.RWMutex
	db map[string]*item
}

func NewStore() *store {
	return &store{
		mu: &sync.RWMutex{},
		db: make(map[string]*item),
	}
}

type item struct {
	mu    *sync.RWMutex
	value []byte
	ttl   int64
}

func (item *item) set(value []byte, ttl int64) {
	item.mu.Lock()
	defer item.mu.Unlock()

	item.value = value
	item.ttl = ttl
}

func (item *item) get() (value []byte, ttl int64) {
	item.mu.RLock()
	defer item.mu.RUnlock()

	return item.value, item.ttl
}

func NewItem(value []byte, ttl int64) *item {
	return &item{
		mu:    &sync.RWMutex{},
		value: value,
		ttl:   ttl,
	}
}

func (s *store) Get(ctx context.Context, key string) ([]byte, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.db[key]
	if !ok {
		return nil, 0, errors.New("key not found")
	}

	value, ttl := item.get()

	return value, ttl, nil
}

func (s *store) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = NewItem(value, ttl)

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
	s.mu.RLock()
	defer s.mu.RUnlock()

	found := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			found++
		}
	}
	return found, nil
}

func (s *store) Expire(ctx context.Context, key string, seconds int64) (int, error) {
	num, err := s.Exists(ctx, []string{key})
	if err != nil {
		return 0, err
	}
	if num == 0 {
		return 0, err
	}

	go time.AfterFunc(time.Duration(seconds)*time.Second, func() {
		s.Del(ctx, []string{key})
	})

	return 1, nil
}

func (s *store) Incr(ctx context.Context, key string) error {
	val, _, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	// todo: there must be a better way to do this, shifting?
	number, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return err
	}

	number = number + 1
	s.Set(ctx, key, []byte(fmt.Sprintf("%d", number)), 0)

	return nil
}

func (s *store) Decr(ctx context.Context, key string) error {
	val, _, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	// todo: there must be a better way to do this, shifting?
	number, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return err
	}

	number = number - 1
	s.Set(ctx, key, []byte(fmt.Sprintf("%d", number)), 0)

	return nil
}
