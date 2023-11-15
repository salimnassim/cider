package cider

import (
	"bytes"
	"math/rand"
	"slices"
	"sync"
	"testing"
	"time"

	"context"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	_, _, err := store.Get(ctx, "notexist")
	if err == nil {
		t.Error(err)
	}
}

func TestSetGet(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	test := []byte("value")

	err := store.Set(ctx, "key", test, time.Now().Unix()+1)
	if err != nil {
		t.Error(err)
	}

	val, _, err := store.Get(ctx, "key")
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(val, test) {
		t.Errorf("bytes are not equal for %v and %v", val, test)
	}

}

func TestSetDel(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	err := store.Set(ctx, "key", []byte("value"), 0)
	if err != nil {
		t.Error(err)
	}

	deletes, err := store.Del(ctx, []string{"key", "nope"})
	if err != nil {
		t.Error(err)
	}

	_, ok := store.db["key"]
	if deletes != 1 || ok {
		t.Errorf("number of deletes should be 1 and not %v", deletes)
	}
}

func TestExists(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	err := store.Set(ctx, "exists", []byte{0x01}, -1)
	if err != nil {
		t.Error(err)
	}

	num, err := store.Exists(ctx, []string{"exists"})
	if err != nil {
		t.Error(err)
	}

	if num != 1 {
		t.Errorf("got: %d, want: %d", num, 1)
	}

	num, err = store.Exists(ctx, []string{"noexists"})
	if err != nil {
		t.Error(err)
	}

	if num != 0 {
		t.Errorf("got: %d, want: %d", num, 0)
	}
}

func TestExpire(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	rep, err := store.Expire(ctx, "test", 100)
	if err != nil {
		t.Error(err)
	}

	if rep != 0 {
		t.Errorf("got: %d, want: %d", rep, 0)
	}

	err = store.Set(ctx, "test2", []byte{0x01}, -1)
	if err != nil {
		t.Error(err)
	}

	rep, err = store.Expire(ctx, "test2", 5)
	if err != nil {
		t.Error(err)
	}

	if rep != 1 {
		t.Errorf("got: %d, want: %d", rep, 1)
	}

}

func TestSetConcurrency(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	var wg sync.WaitGroup
	for i := 1; i < 10000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			keys := []string{"key1", "key2", "key3"}
			values := []string{"val1", "val2", "val3"}

			key := keys[r.Intn(len(keys))]
			value := values[r.Intn(len(values))]

			store.Set(ctx, key, []byte(value), 0)
		}()
	}
	wg.Wait()
}

func TestSetGetConcurrency(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	var wg sync.WaitGroup
	for i := 1; i < 10000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			keys := []string{"key1", "key2", "key3"}
			values := []string{"val1", "val2", "val3"}

			key := keys[r.Intn(len(keys))]
			value := values[r.Intn(len(values))]

			if r.Int()%2 == 0 {
				store.Set(ctx, key, []byte(value), time.Now().Unix()+100)
			} else {
				store.Get(ctx, key)
			}

		}()
	}
	wg.Wait()
}

func TestIncr(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	key := "test1"
	value := "100"

	err := store.Set(ctx, key, []byte(value), time.Now().Unix()+1)
	if err != nil {
		t.Error(err)
	}

	err = store.Incr(ctx, key)
	if err != nil {
		t.Error(err)
	}

	v, _, err := store.Get(ctx, key)
	if err != nil {
		t.Error(err)
	}

	if slices.Compare([]byte{49, 48, 49}, v) != 0 {
		t.Errorf("want: %v, got: %v", []byte{49, 48, 49}, v)
	}
}

func TestIncrIncr(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	key := "test1"
	value := "100"

	err := store.Set(ctx, key, []byte(value), time.Now().Unix()+100)
	if err != nil {
		t.Error(err)
	}

	err = store.Incr(ctx, key)
	if err != nil {
		t.Error(err)
	}

	err = store.Incr(ctx, key)
	if err != nil {
		t.Error(err)
	}

	v, _, err := store.Get(ctx, key)
	if err != nil {
		t.Error(err)
	}

	if slices.Compare([]byte{49, 48, 50}, v) != 0 {
		t.Errorf("want: %v, got: %v", []byte{49, 48, 50}, v)
	}
}

func TestDecr(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	key := "test1"
	value := "100"

	err := store.Set(ctx, key, []byte(value), time.Now().Unix()+100)
	if err != nil {
		t.Error(err)
	}

	err = store.Decr(ctx, key)
	if err != nil {
		t.Error(err)
	}

	v, _, err := store.Get(ctx, key)
	if err != nil {
		t.Error(err)
	}

	if slices.Compare([]byte{57, 57}, v) != 0 {
		t.Errorf("want: %v, got: %v", []byte{57, 57}, v)
	}
}

func TestTTL(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	key := "test1"
	value := "test"
	now := time.Now().Unix() + 100

	err := store.Set(ctx, key, []byte(value), now)
	if err != nil {
		t.Error(err)
	}

	ttl, err := store.TTL(ctx, key)
	if err != nil {
		t.Error(err)
	}

	if ttl != now {
		t.Errorf("want: %v, got %v", now, ttl)
	}

	ttl, err = store.TTL(ctx, "nokey")
	if err != nil {
		t.Error(err)
	}

	if ttl != -2 {
		t.Errorf("want: %d, got %d", -2, ttl)
	}

	err = store.Set(ctx, "existsnoexpire", []byte{0x00}, -1)
	if err != nil {
		t.Error(err)
	}

	ttl, err = store.TTL(ctx, "existsnoexpire")
	if err != nil {
		t.Error(err)
	}

	if ttl != -1 {
		t.Errorf("want: %d, got %d", -1, ttl)
	}

}
