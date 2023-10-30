package cider

import (
	"bufio"
	"context"
	"fmt"
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
			s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
			continue
		}

		switch op.Name {
		case OperationSet:
			err := store.Set(s.ctx, op.Keys[0], []byte(op.Value))
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte("+OK\r\n")
		case OperationGet:
			if len(op.Keys) == 0 {
				s.out <- []byte("-ERR GET value cannot be empty\r\n")
				continue
			}
			res, err := store.Get(s.ctx, op.Keys[0])
			if res == nil {
				s.out <- []byte("_\r\n")
				continue
			}
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(res), res))
		case OperationDel:
			num, err := store.Del(s.ctx, op.Keys)
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte(fmt.Sprintf(":%d\r\n", num))
		case OperationExists:
			num, err := store.Exists(s.ctx, op.Keys)
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte(fmt.Sprintf(":%d\r\n", num))
		case OperationExpire:
			seconds, err := strconv.ParseInt(op.Value, 10, 64)
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			res, err := store.Expire(s.ctx, op.Keys[0], seconds)
			if err != nil {
				// todo: make error readablew
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte(fmt.Sprintf(":%d\r\n", res))
		case OperationIncr:
			err := store.Incr(s.ctx, op.Keys[0])
			if err != nil {
				// todo: make error readablew
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte("+OK\r\n")
			continue
		case OperationDecr:
			err := store.Decr(s.ctx, op.Keys[0])
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte("+OK\r\n")
			continue
		default:
			s.out <- []byte("-ERR unknown command\r\n")
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
