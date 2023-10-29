package main

import (
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/salimnassim/cider"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	address, ok := os.LookupEnv("ADDRESS")
	if !ok {
		log.Fatal().Msg("unable to read environment variable ADDRESS")
	}

	store := cider.NewStore()

	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal().Err(err).Msg("unable to listen")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}

		session := cider.NewSession(conn)
		go session.HandleOut()
		go session.HandleIn(store)
	}

}
