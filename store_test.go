package cider

import (
	"bytes"
	"testing"

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

	err := store.Set(ctx, "key", []byte("value"))
	if err != nil {
		t.Error(err)
	}

	val, err := store.Get(ctx, "key")
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(val, []byte("value")) {
		t.Errorf("bytes are not equal for %v and %v", val, []byte("value"))
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
