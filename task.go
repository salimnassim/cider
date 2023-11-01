package cider

import (
	"time"
)

type task struct {
	interval time.Duration
	function func(Storer)
	store    Storer
	stop     chan bool
}

func NewTask(interval time.Duration, f func(Storer), s Storer) *task {
	return &task{
		interval: interval,
		function: f,
		store:    s,
		stop:     make(chan bool),
	}
}

func (t *task) Run() {
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
