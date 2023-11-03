package cider

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func TestOk(t *testing.T) {
	want := []byte("+OK\r\n")
	res := replyOK()
	if slices.Compare(res, want) != 0 {
		t.Errorf("want: %v, got %v", want, res)
	}
}

func TestError(t *testing.T) {
	want := []byte("-ERR test\r\n")
	err := errors.New("test")
	res := replyError(
		err,
	)
	if slices.Compare(res, want) != 0 {
		t.Errorf("want: %v, got %v", want, res)
	}
}

func TestNil(t *testing.T) {
	want := []byte("_\r\n")
	res := replyNil()
	if slices.Compare(res, want) != 0 {
		t.Errorf("want: %v, got %v", want, res)
	}
}

func TestString(t *testing.T) {
	want := []byte(fmt.Sprintf("$%d\r\n%s\r\n", 3, "foo"))
	res := replyString([]byte("foo"))
	if slices.Compare(res, want) != 0 {
		t.Errorf("want: %v, got %v", want, res)
	}
}
func TestInteger(t *testing.T) {
	want := []byte(fmt.Sprintf(":%d\r\n", 42))
	res := replyInteger(42)
	if slices.Compare(res, want) != 0 {
		t.Errorf("want: %v, got %v", want, res)
	}
}
