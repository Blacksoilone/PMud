package rawterm_test

import (
	"errors"
	"testing"

	"PMud/internal/client/rawterm"
)

var errMakeRaw = errors.New("make raw failed")
var errRestore = errors.New("restore failed")

type fakeController struct {
	isTerminal   bool
	makeRawError error
	restoreError error
	makeRawCalls int
	restoreCalls int
	restoredFD   int
}

func (f *fakeController) IsTerminal(int) bool {
	return f.isTerminal
}

func (f *fakeController) MakeRaw(int) (*rawterm.State, error) {
	f.makeRawCalls++
	if f.makeRawError != nil {
		return nil, f.makeRawError
	}
	return &rawterm.State{}, nil
}

func (f *fakeController) Restore(fd int, state *rawterm.State) error {
	f.restoreCalls++
	f.restoredFD = fd
	if state == nil {
		return errors.New("missing state")
	}
	return f.restoreError
}

func (f *fakeController) Size(int) (int, int, error) {
	return 0, 0, nil
}

func TestStartFailsWhenFileDescriptorIsNotTerminal(t *testing.T) {
	controller := &fakeController{}

	session, err := rawterm.Start(7, controller)

	if !errors.Is(err, rawterm.ErrNotTerminal) {
		t.Fatalf("Start error = %v, want ErrNotTerminal", err)
	}
	if session != nil {
		t.Fatalf("session = %#v, want nil", session)
	}
	if controller.makeRawCalls != 0 {
		t.Fatalf("MakeRaw calls = %d, want 0", controller.makeRawCalls)
	}
}

func TestSessionCloseRestoresTerminalOnce(t *testing.T) {
	controller := &fakeController{isTerminal: true}
	session, err := rawterm.Start(7, controller)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	if err := session.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := session.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	if controller.makeRawCalls != 1 {
		t.Fatalf("MakeRaw calls = %d, want 1", controller.makeRawCalls)
	}
	if controller.restoreCalls != 1 {
		t.Fatalf("Restore calls = %d, want 1", controller.restoreCalls)
	}
	if controller.restoredFD != 7 {
		t.Fatalf("restored fd = %d, want 7", controller.restoredFD)
	}
}

func TestStartReturnsMakeRawError(t *testing.T) {
	controller := &fakeController{isTerminal: true, makeRawError: errMakeRaw}

	session, err := rawterm.Start(7, controller)

	if !errors.Is(err, errMakeRaw) {
		t.Fatalf("Start error = %v, want %v", err, errMakeRaw)
	}
	if session != nil {
		t.Fatalf("session = %#v, want nil", session)
	}
	if controller.restoreCalls != 0 {
		t.Fatalf("Restore calls = %d, want 0", controller.restoreCalls)
	}
}

func TestCloseReturnsRestoreError(t *testing.T) {
	controller := &fakeController{isTerminal: true, restoreError: errRestore}
	session, err := rawterm.Start(7, controller)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	err = session.Close()

	if !errors.Is(err, errRestore) {
		t.Fatalf("Close error = %v, want %v", err, errRestore)
	}
	if controller.restoreCalls != 1 {
		t.Fatalf("Restore calls = %d, want 1", controller.restoreCalls)
	}
}
