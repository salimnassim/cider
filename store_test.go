package cider

import (
	"bytes"
	"math/rand"
	"sync"
	"testing"
	"time"

	"context"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	_, err := store.Get(ctx, "notexist")
	if err == nil {
		t.Error(err)
	}
}

func TestSetGet(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	test := []byte("value")

	err := store.Set(ctx, "key", test)
	if err != nil {
		t.Error(err)
	}

	val, err := store.Get(ctx, "key")
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

	err := store.Set(ctx, "key", []byte("value"))
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

			store.Set(ctx, key, []byte(value))
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
				store.Set(ctx, key, []byte(value))
			} else {
				store.Get(ctx, key)
			}

		}()
	}
	wg.Wait()
}
