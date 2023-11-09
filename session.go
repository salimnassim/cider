package cider

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"

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

		switch t := op.(type) {
		case opSet:
			err := store.Set(s.ctx, t.key, t.value, 0)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyOK()
		case opGet:
			value, _, err := store.Get(s.ctx, t.key)
			if err != nil && err.Error() == "key not found" {
				s.out <- replyNil()
				continue
			}
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyString(value)
		case opDel:
			num, err := store.Del(s.ctx, t.keys)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(num)
		case opExists:
			num, err := store.Exists(s.ctx, t.keys)
			if err != nil {
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(num)
		case opExpire:
			res, err := store.Expire(s.ctx, t.key, t.ttl)
			if err != nil {
				// todo: make error readable
				s.out <- replyError(err)
				continue
			}
			s.out <- replyInteger(res)
		case opIncr:
			err := store.Incr(s.ctx, t.key)
			if err != nil {
				// todo: make error readable
				s.out <- replyError(err)
				continue
			}
			s.out <- replyOK()
			continue
		case opDecr:
			err := store.Decr(s.ctx, t.key)
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
