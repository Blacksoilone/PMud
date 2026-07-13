# Client Raw Byte Input Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Connect terminal input bytes to the existing `keyinput.Decode` and `TUIRuntime.ApplyInput` path.

**Architecture:** Keep terminal raw-mode lifecycle out of this slice. Add a client-level forwarding function that reads byte chunks from an `io.Reader`, decodes them with `internal/client/keyinput`, applies each decoded action to `TUIRuntime`, stops cleanly on quit, and propagates read/runtime errors.

**Tech Stack:** Go standard library, existing `internal/client`, `internal/client/keyinput`, `internal/client/tui`, and `internal/client/screen` packages.

---

## File Structure

- Modify `internal/client/client.go`: add `ForwardTUIKeyInput`.
- Modify `internal/client/client_test.go`: add raw byte input loop tests.

## Task 1: Raw Byte Input Loop Tests

- [ ] Add tests to `internal/client/client_test.go`.

Required behavior:

- ASCII and CJK bytes update the TUI prompt through `TUIRuntime.ApplyInput`.
- Backspace removes the last rune before submit.
- Enter submits the resolved command to the server.
- Ctrl+U clears the prompt.
- Ctrl+C stops the loop without error and ignores later bytes.

Run:

```bash
go test ./internal/client -run 'TestForwardTUIKeyInput' -count=1
```

Expected before implementation: compile failure for missing `ForwardTUIKeyInput`.

## Task 2: Implement Raw Byte Input Loop

- [ ] Add function to `internal/client/client.go`:

```go
func ForwardTUIKeyInput(input io.Reader, server io.Writer, runtime *TUIRuntime) error
```

Implementation requirements:

- read from `input` into a small buffer;
- for each chunk, call `keyinput.Decode(chunk)`;
- if action has `Quit`, return nil immediately;
- otherwise call `runtime.ApplyInput(action.Input, server)`;
- EOF returns nil;
- read errors and runtime errors are returned.

Run:

```bash
go test ./internal/client -run 'TestForwardTUIKeyInput' -count=1
```

Expected after implementation: pass.

## Task 3: Verification and Commit

- [ ] Run `gofmt -w internal/client/client.go internal/client/client_test.go`.
- [ ] Run `go test ./internal/client -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client`.
- [ ] Measure pure LOC for changed Go files.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `接入TUI按键输入循环`.
