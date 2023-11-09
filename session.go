package cider

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strconv"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Session struct {
	id     uuid.UUID
	conn   net.Conn
	reader io.Reader
	ctx    context.Context
	in     chan []byte
	out    chan []byte
	stop   chan bool
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		id:     uuid.New(),
		conn:   conn,
		ctx:    context.Background(),
		reader: bufio.NewReader(conn),
		in:     make(chan []byte, 1),
		out:    make(chan []byte, 1),
		stop:   make(chan bool),
	}
}

func (s *Session) HandleIn(store Storer) {
	scanner := bufio.NewScanner(s.reader)

	for scanner.Scan() {
		bytes := scanner.Bytes()

		op, err := ParseCommand(bytes)
		if err != nil {
			s.out <- replyError(err)
			continue
		}

		switch op.name {
		case OperationSet:
			err := store.Set(s.ctx, op.keys[0], []byte(op.value), 0)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyOK()
		case OperationGet:
			if len(op.keys) == 0 {
				s.out <- replyError(errors.New("value cannot be empty"))
				continue
			}
			value, _, err := store.Get(s.ctx, op.keys[0])
			if err != nil && err.Error() == "key not found" {
				s.out <- replyNil()
				continue
			}
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyString(value)
		case OperationDel:
			num, err := store.Del(s.ctx, op.keys)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(num)
		case OperationExists:
			num, err := store.Exists(s.ctx, op.keys)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(num)
		case OperationExpire:
			seconds, err := strconv.ParseInt(op.value, 10, 64)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			res, err := store.Expire(s.ctx, op.keys[0], seconds)
			if err != nil {
				// todo: make error readable
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(res)
		case OperationIncr:
			err := store.Incr(s.ctx, op.keys[0])
			if err != nil {
				// todo: make error readable
				s.out <- replyError(err)
				continue
			}
			s.out <- replyOK()
			continue
		case OperationDecr:
			err := store.Decr(s.ctx, op.keys[0])
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyOK()
			continue
		default:
			s.out <- replyError(errors.New("unknown command"))
			continue
		}
	}
}

func (s *Session) HandleOut() {
	for message := range s.out {
		_, err := s.conn.Write(message)
		if err != nil {
			log.Error().Err(err).Msgf("cant write message to session %s", s.id)
			break
		}
	}

	err := s.conn.Close()
	if err != nil {
		log.Error().Err(err).Msg("error closing connection")
	}
}
