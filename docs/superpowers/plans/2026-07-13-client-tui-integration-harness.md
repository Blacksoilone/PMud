# Client TUI Integration Harness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the pure `internal/client/tui` runtime skeleton into the existing client protocol flow behind an explicit `mudclient --tui` mode.

**Architecture:** Keep the existing line-based client as the default stable path. Add a TUI-aware renderer in `internal/client` that parses server protocol lines, observes them with `client.State`, updates `tui.Model`, and writes `tui.View(...).String()` after each event. Add a small `--tui` switch in `cmd/mudclient` that uses this renderer without introducing raw terminal mode yet.

**Tech Stack:** Go standard library, existing `internal/client`, `internal/client/tui`, `internal/content`, and `internal/protocol` packages.

---

## File Structure

- Modify `internal/client/client.go`: add TUI protocol rendering functions next to existing line renderers.
- Modify `internal/client/client_test.go`: add TUI renderer tests using the tutorial catalog and state.
- Modify `cmd/mudclient/main.go`: parse `--tui` and optional address, select renderer mode.

## Task 1: TUI Protocol Renderer Tests

- [ ] Add tests to `internal/client/client_test.go` before implementation.

Required tests:

```go
func TestRenderTUIProtocolLinesRendersObservedEvents(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatalf("Compile TutorialSource: %v", err)
	}
	state := client.NewState(compiled.Client)
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = client.RenderTUIProtocolLines(input, &output, state, 48, 3)

	if err != nil {
		t.Fatalf("RenderTUIProtocolLines: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, "练习场入口") {
		t.Fatalf("output does not include room name:\n%s", got)
	}
	if !strings.Contains(got, "> ") {
		t.Fatalf("output does not include prompt:\n%s", got)
	}
}
```

```go
func TestRenderTUIObservedProtocolLinesUpdatesCommandResolution(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatalf("Compile TutorialSource: %v", err)
	}
	state := client.NewState(compiled.Client)
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = client.RenderTUIObservedProtocolLines(input, &output, state, 48, 3)

	if err != nil {
		t.Fatalf("RenderTUIObservedProtocolLines: %v", err)
	}
	line := state.ResolveCommand("get 旧油灯")
	if line != "get item.tutorial.old_lantern" {
		t.Fatalf("resolved command = %q", line)
	}
}
```

- [ ] Run `go test ./internal/client -run 'TestRenderTUI' -count=1`.

Expected before implementation: compile failure for missing `RenderTUIProtocolLines` and `RenderTUIObservedProtocolLines`.

## Task 2: Implement TUI Protocol Renderer

- [ ] Add functions to `internal/client/client.go`:

```go
func RenderTUIProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error
func RenderTUIObservedProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error
```

Implementation requirements:

- scan input line-by-line like the existing renderer;
- parse with `protocol.ParseLine`;
- call `state.Observe(event)` for observed mode;
- update `tui.Model` with `tui.ApplyEvent`;
- write `tui.View(model, state.catalog, width).String()` after each event;
- preserve `ErrProtocolLine` wrapping behavior on parse errors.

- [ ] Run `go test ./internal/client -run 'TestRenderTUI' -count=1`.

Expected after implementation: pass.

## Task 3: Add `mudclient --tui` Wiring

- [ ] Modify `cmd/mudclient/main.go` so `--tui` selects the TUI renderer.

Behavior:

- `mudclient` keeps the existing line-rendered mode.
- `mudclient 127.0.0.1:4000` keeps the existing address override.
- `mudclient --tui` uses default address and TUI renderer.
- `mudclient --tui 127.0.0.1:4000` uses TUI renderer and address override.

Implementation can use a small local argument parser in `main.go`; do not introduce a CLI framework.

- [ ] Run `go test ./cmd/mudclient -count=1` and `go test ./... -count=1`.

Expected after implementation: pass.

## Task 4: Verification and Commit

- [ ] Run `gofmt -w internal/client/client.go internal/client/client_test.go cmd/mudclient/main.go`.
- [ ] Run `go test ./internal/client -count=1`.
- [ ] Run `go test ./cmd/mudclient -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on modified Go files.
- [ ] Measure pure LOC for modified Go files and keep each under 250.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage only this slice and commit with `接入客户端TUI渲染模式`.
