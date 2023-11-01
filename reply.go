package cider

import "fmt"

type Replyer interface {
	replyOK() []byte
	replyError(err error) []byte
	replyString(value []byte) []byte
	replyInteger(value int64) []byte
}

func replyOK() []byte {
	return []byte("+OK")
}

func replyError(err error) []byte {
	return []byte(fmt.Sprintf("-ERR %s", err.Error()))
}

func replyString(value []byte) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}

func replyInteger(value int64) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", value))
}
