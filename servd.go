package servd

import (
	"context"
	"errors"
	"fmt"
)

// Status type of the service. It describe the lifecycle of the service.
type Status int

const (
	// Created status. Phase when the service just created not started yet.
	Created Status = iota

	// Running status. Phase when the service is running.
	Running

	// Stopped status. Phase when the service is stopped.
	Stopped
)

var statusNames = []string{
	"Created",
	"Running",
	"Stopped",
}

// Name of the status.
func (s Status) Name() string {
	return statusNames[s]
}

func (s Status) String() string {
	return s.Name()
}

// Handler is the application handler.
//
// Servd will pass ctx as cancellable context.Context to Handle(...) method. Listen to the ctx.Done() to as shutdown signal.
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
	statSubscribers map[Status][]chan<- Status
	status          Status
	stop            func()
}

// Run the service. When Run failed or error occur and forced the service to stop,
// then it should return error.
func (s *Servd) Run() error {
	if s.Handler == nil {
		return errors.New("servd: no handler")
	}

	if s.status > Created {
		return fmt.Errorf("servd: cannot run in status '%s'", s.status)
	}

	s.changeStatus(Running)
	defer s.changeStatus(Stopped)
	ctx, stop := context.WithCancel(context.Background())
	s.stop = stop
	return s.Handler.Handle(ctx)
}

// Status of the service.
func (s *Servd) Status() Status {
	return s.status
}

// WaitForStatus waiting for specific status.
//
// It will wait until it reach or passed the expected status.
// If the status has reach the one that it wait for, then it will return immediately.
// If the status has pass the one that it wait for, then it will return the latest one.
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
// Return false if the Stop already been called or service already stopped.
func (s *Servd) Stop() bool {
	if s.status == Created {
		s.changeStatus(Stopped)
		return true
	}

	if s.stop == nil && s.status == Running {
		return false
	}

	s.stop()
	return true
}

// StopWait wills top and wait.
func (s *Servd) StopWait(ctx context.Context) error {
	_, err := s.WaitForStatus(ctx, Stopped)
	return err
}

func (s *Servd) notifyStatus(c chan<- Status, stat Status) {
	if s.statSubscribers == nil {
		s.statSubscribers = make(map[Status][]chan<- Status)
	}

	s.statSubscribers[stat] = append(s.statSubscribers[stat], c)
}

func (s *Servd) cancelNotifyStatus(c chan<- Status, stat Status) {
	if s.statSubscribers == nil {
		return
	}

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

	if s.statSubscribers == nil {
		return
	}

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
