package cider

import "fmt"

type Replyer interface {
	replyOK() []byte
	replyError(err error) []byte
	replyNil() []byte
	replyString(value []byte) []byte
	replyInteger(value int64) []byte
}

func replyOK() []byte {
	return []byte("+OK\r\n")
}

func replyError(err error) []byte {
	return []byte(fmt.Sprintf("-ERR %s\r\n", err.Error()))
}

func replyNil() []byte {
	return []byte("_\r\n")
}

func replyString(value []byte) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}

func replyInteger(value int64) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", value))
}
