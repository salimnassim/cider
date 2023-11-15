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
	Del(ctx context.Context, keys []string) (deleted int64, err error)
	// Checks if keys exist in database. Returns the number of keys found.
	Exists(ctx context.Context, keys []string) (found int64, err error)
	// Expires a key after n seconds.
	Expire(ctx context.Context, key string, ttl int64) (result int64, err error)
	// Increments a key.
	Incr(ctx context.Context, key string) (err error)
	// Decrements a key.
	Decr(ctx context.Context, key string) (err error)
	// Gets the TTL of a key. -2 if it does not exist or -1 if key exists but no TTL is set.
	TTL(ctx context.Context, key string) (result int64, err error)
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

func (item *item) setValue(value []byte) {
	item.mu.Lock()
	defer item.mu.Unlock()

	item.value = value
}

func (item *item) setTTL(ttl int64) {
	item.mu.Lock()
	defer item.mu.Unlock()

	item.ttl = ttl
}

// Returns value and ttl as copies.
func (item *item) get() (value []byte, ttl int64) {
	item.mu.RLock()
	defer item.mu.RUnlock()

	return item.value, item.ttl
}

func NewItem(value []byte, ttl int64) *item {

	if ttl == 0 {
		ttl = -1
	}

	return &item{
		mu:    &sync.RWMutex{},
		value: value,
		ttl:   ttl,
	}
}

func (s *store) Get(ctx context.Context, key string) ([]byte, int64, error) {
	s.mu.RLock()
	item, ok := s.db[key]
	if !ok {
		return nil, 0, errors.New("key not found")
	}
	s.mu.RUnlock()

	value, ttl := item.get()

	if ttl != -1 && ttl <= time.Now().Unix() {
		defer s.Del(ctx, []string{key})
		return []byte{}, 0, errors.New("key not found")
	}

	return value, ttl, nil
}

func (s *store) Set(ctx context.Context, key string, value []byte, ttl int64) error {
	item := NewItem(value, ttl)

	s.mu.Lock()
	s.db[key] = item
	s.mu.Unlock()

	return nil
}

func (s *store) Del(ctx context.Context, keys []string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	deletes := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			delete(s.db, key)
			deletes++
		}
	}
	return int64(deletes), nil
}

func (s *store) Exists(ctx context.Context, keys []string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	found := 0
	for _, key := range keys {
		if _, ok := s.db[key]; ok {
			found++
		}
	}
	return int64(found), nil
}

func (s *store) Expire(ctx context.Context, key string, seconds int64) (int64, error) {
	num, err := s.Exists(ctx, []string{key})
	if err != nil {
		return 0, err
	}
	if num == 0 {
		return 0, err
	}

	s.mu.RLock()
	s.db[key].setTTL(time.Now().Unix() + seconds)
	s.mu.RUnlock()

	return 1, nil
}

func (s *store) Incr(ctx context.Context, key string) error {
	val, ttl, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	// todo: there must be a better way to do this, shifting?
	number, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return err
	}

	number = number + 1
	s.Set(ctx, key, []byte(fmt.Sprintf("%d", number)), ttl)

	return nil
}

func (s *store) Decr(ctx context.Context, key string) error {
	val, ttl, err := s.Get(ctx, key)
	if err != nil {
		return err
	}

	// todo: there must be a better way to do this, shifting?
	number, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		return err
	}

	number = number - 1
	s.Set(ctx, key, []byte(fmt.Sprintf("%d", number)), ttl)

	return nil
}

func (s *store) TTL(ctx context.Context, key string) (int64, error) {
	_, ttl, err := s.Get(ctx, key)
	if err != nil && err.Error() == "key not found" {
		// The command returns -2 if the key does not exist.
		return -2, nil
	}
	if err != nil {
		return 0, err
	}
	if ttl == 0 {
		// The command returns -1 if the key exists but has no associated expire.
		return -1, nil
	}
	return ttl, nil
}
