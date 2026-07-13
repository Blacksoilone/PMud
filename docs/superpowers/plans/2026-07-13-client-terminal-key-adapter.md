# Client Terminal Key Adapter Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a pure terminal key adapter that converts input bytes into `tui.Input` actions or a quit signal.

**Architecture:** Keep terminal raw mode and runtime wiring out of this slice. The adapter is a small pure package under `internal/client/keyinput`; it decodes complete UTF-8 input chunks and known control bytes into typed actions that `TUIRuntime.ApplyInput` can consume in a later slice.

**Tech Stack:** Go standard library, existing `internal/client/tui` package.

---

## File Structure

- Create `internal/client/keyinput/keyinput.go`: action type and decoder.
- Create `internal/client/keyinput/keyinput_test.go`: ASCII, CJK, and control-key tests.

## Task 1: Decoder Tests

- [ ] Add `internal/client/keyinput/keyinput_test.go` first.

Required tests:

- ASCII text `get` decodes into three `InputText` actions.
- CJK text `旧灯` decodes into two `InputText` actions with whole runes.
- Enter as `\n` and `\r` decodes into `InputSubmit`.
- Backspace as `0x7f` and `\b` decodes into `InputBackspace`.
- Ctrl+U (`0x15`) decodes into `InputClear`.
- Ctrl+C (`0x03`) decodes into a quit action.

Run:

```bash
go test ./internal/client/keyinput -count=1
```

Expected before implementation: package or symbol compile failure.

## Task 2: Implement Decoder

- [ ] Add `internal/client/keyinput/keyinput.go`.

Required shape:

```go
package keyinput

import "PMud/internal/client/tui"

type Action struct {
	Input tui.Input
	Quit  bool
}

func Decode(data []byte) []Action
```

Implementation requirements:

- decode UTF-8 runes from `data`;
- regular printable runes become `Action{Input: tui.Input{Kind: tui.InputText, Text: string(r)}}`;
- `\n` and `\r` become submit;
- `0x7f` and `\b` become backspace;
- Ctrl+U becomes clear;
- Ctrl+C becomes `Action{Quit: true}`;
- ignore invalid UTF-8 bytes in this slice rather than emitting replacement characters;
- do not parse escape sequences or arrow keys yet.

Run:

```bash
go test ./internal/client/keyinput -count=1
```

Expected after implementation: pass.

## Task 3: Verification and Commit

- [ ] Run `gofmt -w internal/client/keyinput`.
- [ ] Run `go test ./internal/client/keyinput -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client/keyinput`.
- [ ] Measure pure LOC for new Go files.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `新增终端按键解码器`.
