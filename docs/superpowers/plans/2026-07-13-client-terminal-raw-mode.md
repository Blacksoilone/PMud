# Client Terminal Raw Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `mudclient --tui` enter terminal raw mode, read key bytes through the existing TUI key-input path, and restore the terminal on exit.

**Architecture:** Add a small `internal/client/rawterm` package that owns raw-mode lifecycle. Keep it testable with an injectable controller, then use the real `golang.org/x/term` controller from `cmd/mudclient`. Non-TUI mode stays line based.

**Tech Stack:** Go standard library, `golang.org/x/term`, existing `internal/client` TUI runtime and key-input pipeline.

---

## File Structure

- Create `internal/client/rawterm/rawterm.go`: raw-mode session lifecycle and real terminal controller.
- Create `internal/client/rawterm/rawterm_test.go`: fake-controller lifecycle tests.
- Modify `cmd/mudclient/main.go`: in `--tui`, start raw mode and use `client.ForwardTUIKeyInput`.
- Modify `go.mod` / `go.sum`: add `golang.org/x/term` dependency.

## Task 1: Raw Session Lifecycle Tests

- [ ] Add `internal/client/rawterm/rawterm_test.go` first.

Required behavior:

- `Start` fails when the fd is not a terminal and does not call `MakeRaw`.
- `Start` calls `MakeRaw` once and `Close` calls `Restore` once.
- `Close` is idempotent and does not restore twice.
- `MakeRaw` error is returned.
- `Restore` error is returned from `Close`.

Run:

```bash
go test ./internal/client/rawterm -count=1
```

Expected before implementation: package or symbol compile failure.

## Task 2: Implement Raw Session

- [ ] Add `internal/client/rawterm/rawterm.go`.

Required shape:

```go
package rawterm

type State struct{}

type Controller interface {
	IsTerminal(fd int) bool
	MakeRaw(fd int) (*State, error)
	Restore(fd int, state *State) error
}

type Session struct { ... }

func Start(fd int, controller Controller) (*Session, error)
func (s *Session) Close() error
func RealController() Controller
```

Implementation requirements:

- return `ErrNotTerminal` when `IsTerminal` is false;
- call `MakeRaw` only after terminal check succeeds;
- restore exactly once;
- `RealController` adapts `golang.org/x/term` state to this package's `State` wrapper;
- do not call `os.Exit` or manipulate stdin/stdout directly in this package.

Run:

```bash
go test ./internal/client/rawterm -count=1
```

Expected after implementation: pass.

## Task 3: Wire `mudclient --tui`

- [ ] Modify `cmd/mudclient/main.go`.

Required behavior:

- non-TUI mode still uses `client.ForwardResolvedCommands`;
- TUI mode starts raw mode on `os.Stdin.Fd()`;
- if raw mode starts, defer restore before launching input forwarding;
- TUI input goroutine uses `client.ForwardTUIKeyInput(os.Stdin, conn, runtime)`;
- if raw mode cannot start, return a clear fatal error.

Run:

```bash
go test ./cmd/mudclient -count=1
go test ./internal/client/rawterm ./internal/client ./cmd/mudclient -count=1
```

Expected after wiring: pass.

## Task 4: Verification and Commit

- [ ] Run `gofmt -w internal/client/rawterm cmd/mudclient/main.go`.
- [ ] Run `go test ./internal/client/rawterm ./internal/client ./cmd/mudclient -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client/rawterm` and `cmd/mudclient`.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `接入TUI终端原始模式`.
