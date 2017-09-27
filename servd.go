package servd

import (
	"context"
	"errors"
)

// Status type of the service.
type Status int

const (
	// Created status.
	Created Status = iota

	// Running status.
	Running

	// Stopped status
	Stopped
)

// Handler is the application handler.
type Handler interface {
	Handle(context.Context) error
}

// HandleFunc is the Handler adapter.
type HandleFunc func(context.Context) error

// Handle the application run.
func (f HandleFunc) Handle(ctx context.Context) error {
	return f(ctx)
}

// Servd is the service daemon.
type Servd struct {
	Handler         Handler
	status          Status
	statSubscribers map[Status][]chan<- Status
	stop            func()
}

// Run the service.
func (s *Servd) Run() error {
	if s.Handler == nil {
		return errors.New("servd: no handler")
	}

	if s.status > Created {
		return errors.New("servd: cannot re-run")
	}

	// initialize
	s.statSubscribers = make(map[Status][]chan<- Status)

	s.changeStatus(Running)
	defer s.changeStatus(Stopped)

	ctx, cancel := context.WithCancel(context.Background())
	s.stop = cancel
	return s.handle(ctx)
}

func (s *Servd) handle(ctx context.Context) error {
	return s.Handler.Handle(ctx)
}

// Status of the service.
func (s *Servd) Status() Status {
	return s.status
}

// WaitForStatus waiting for specific status.
// It will wait until it reach or passed the expected status.
func (s *Servd) WaitForStatus(ctx context.Context, stat Status) (Status, error) {
	if s.status >= stat {
		return s.status, nil
	}

	c := make(chan Status, 1)
	s.notifyStatus(c, stat)
	defer s.cancelNotifyStatus(c, stat)

	select {
	case <-ctx.Done():
		return Created, ctx.Err()
	case <-c:
		return stat, nil
	}
}

// Stop the service.
// Return true if the service stop has been called or the app is shutting down.
func (s *Servd) Stop() bool {
	if s.stop == nil && s.status == Running {
		return false
	}

	s.stop()
	return true
}

// StopAndWait the service to be stopped.
func (s *Servd) StopAndWait(ctx context.Context) (Status, error) {
	s.Stop()
	return s.WaitForStatus(ctx, Stopped)
}

func (s *Servd) notifyStatus(c chan<- Status, stat Status) {
	s.statSubscribers[stat] = append(s.statSubscribers[stat], c)
}

func (s *Servd) cancelNotifyStatus(c chan<- Status, stat Status) {
	subs := s.statSubscribers[stat]
	i := chanIndex(subs, c)
	if i < 0 {
		return
	}

	// remove value on index i
	s.statSubscribers[stat] = append(subs[:i], subs[i+1:]...)
}

func (s *Servd) changeStatus(stat Status) {
	s.status = stat

	// notifies all subscribers of s
	subs, ok := s.statSubscribers[stat]
	if !ok {
		return
	}

	notify := func(sub chan<- Status, stat Status) {
		sub <- stat
	}
	for _, sub := range subs {
		go notify(sub, stat)
	}
}

func chanIndex(s []chan<- Status, c chan<- Status) int {
	for i, v := range s {
		if v == c {
			return i
		}
	}
	return -1
}