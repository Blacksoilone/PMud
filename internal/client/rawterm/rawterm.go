package rawterm

import (
	"errors"

	"golang.org/x/term"
)

var ErrNotTerminal = errors.New("not a terminal")

type State struct {
	termState *term.State
}

type Controller interface {
	IsTerminal(fd int) bool
	MakeRaw(fd int) (*State, error)
	Restore(fd int, state *State) error
}

type Session struct {
	fd         int
	state      *State
	controller Controller
	closed     bool
}

func Start(fd int, controller Controller) (*Session, error) {
	if !controller.IsTerminal(fd) {
		return nil, ErrNotTerminal
	}
	state, err := controller.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return &Session{fd: fd, state: state, controller: controller}, nil
}

func (s *Session) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	return s.controller.Restore(s.fd, s.state)
}

type realController struct{}

func RealController() Controller {
	return realController{}
}

func (realController) IsTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

func (realController) MakeRaw(fd int) (*State, error) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return &State{termState: state}, nil
}

func (realController) Restore(fd int, state *State) error {
	return term.Restore(fd, state.termState)
}
