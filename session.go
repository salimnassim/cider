package cider

import (
	"bufio"
	"context"
	"fmt"
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

		op := ParseCommand(bytes)

		switch op.Name {
		case OperationSet:
			err := store.Set(s.ctx, op.Keys[0], []byte(op.Value))
			if err != nil {
				s.out <- []byte(fmt.Sprintf("-ERR %s\r\n", err))
				continue
			}
			s.out <- []byte("+OK\r\n")
		case OperationGet:
			res, err := store.Get(s.ctx, op.Keys[0])
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
