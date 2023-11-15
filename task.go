package cider

import (
	"time"

	"github.com/rs/zerolog/log"
)

type task struct {
	name     string
	interval time.Duration
	function func(Storer)
	store    Storer
	stop     chan bool
}

func NewTask(name string, interval time.Duration, f func(Storer), s Storer) *task {
	return &task{
		name:     name,
		interval: interval,
		function: f,
		store:    s,
		stop:     make(chan bool),
	}
}

func (t *task) Run() {
	log.Info().Msgf("starting task %s", t.name)

	ticker := time.NewTicker(t.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				t.function(t.store)
			case <-t.stop:
				ticker.Stop()
				return
			}
		}
	}()
}
