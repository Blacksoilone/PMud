# Client TUI Line Input Runtime Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `mudclient --tui` display submitted input lines in the TUI prompt area before sending them to the server, then clear the prompt after submit.

**Architecture:** Add a small serialized `client.TUIRuntime` that owns the live `tui.Model`, the client `State`, screen redraw, and a mutex. Server events and input submissions both pass through this runtime so redraws do not interleave. Keep input line-based in this slice; do not add raw mode, per-key editing, terminal restore, or a frame ticker.

**Tech Stack:** Go standard library, existing `internal/client`, `internal/client/tui`, `internal/client/screen`, `internal/protocol`, and `cmd/mudclient`.

---

## File Structure

- Create `internal/client/tui_runtime.go`: serialized runtime for event and submitted-line redraws.
- Create `internal/client/tui_runtime_test.go`: runtime behavior tests.
- Modify `internal/client/client.go`: use `TUIRuntime` in TUI protocol renderer and add line-submission forwarding.
- Modify `internal/client/client_test.go`: verify TUI line submission writes the resolved command and redraws input/cleared prompt.
- Modify `cmd/mudclient/main.go`: route `--tui` input through the TUI line-submission path.

## Task 1: Runtime Tests

- [ ] Add `internal/client/tui_runtime_test.go` first.

Required tests:

```go
func TestTUIRuntimeObserveEventRedrawsScreen(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ObserveEvent(protocol.Event{Name: "inventory", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})

	if err != nil {
		t.Fatalf("ObserveEvent: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, screen.FullRedrawPrefix) || !strings.Contains(got, "你带着: 旧油灯") {
		t.Fatalf("output missing redraw or event text:\n%s", got)
	}
}
```

```go
func TestTUIRuntimeSubmitLineRedrawsInputThenClearsPrompt(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{Name: "room", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 48, HistoryLimit: 3})

	err = runtime.SubmitLine("get 旧油灯", &serverOutput)

	if err != nil {
		t.Fatalf("SubmitLine: %v", err)
	}
	if serverOutput.String() != "get item.tutorial.old_lantern\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "> get 旧油灯") || !strings.Contains(got, "| >") {
		t.Fatalf("output missing submitted or cleared prompt:\n%s", got)
	}
}
```

- [ ] Run `go test ./internal/client -run 'TestTUIRuntime' -count=1`.

Expected before implementation: compile failure for missing runtime symbols.

## Task 2: Implement TUIRuntime

- [ ] Create `internal/client/tui_runtime.go`.

Required shape:

```go
type TUIRuntimeConfig struct {
	State        *State
	Output       io.Writer
	Width        int
	HistoryLimit int
}

type TUIRuntime struct { ... }

func NewTUIRuntime(config TUIRuntimeConfig) *TUIRuntime
func (r *TUIRuntime) RenderEvent(event protocol.Event) error
func (r *TUIRuntime) ObserveEvent(event protocol.Event) error
func (r *TUIRuntime) SubmitLine(line string, server io.Writer) error
```

Implementation requirements:

- guard all model mutations and draws with one mutex;
- `RenderEvent` appends the event and redraws without observing item names;
- `ObserveEvent` observes item names before appending and redraws;
- `SubmitLine` applies `tui.InputText`, redraws visible input, applies `tui.InputSubmit`, resolves the command via `State.ResolveCommand`, writes the resolved line plus newline to `server`, then redraws the cleared prompt;
- whitespace-only submissions must not write to `server`, but still clear and redraw consistently.

- [ ] Run `go test ./internal/client -run 'TestTUIRuntime' -count=1`.

Expected after implementation: pass.

## Task 3: Refactor TUI Protocol Renderer Through Runtime

- [ ] Modify `internal/client/client.go`.

Implementation requirements:

- `renderTUIProtocolLines` creates one `TUIRuntime`;
- on each parsed event, call `runtime.ObserveEvent(event)` for observed mode and `runtime.RenderEvent(event)` otherwise;
- preserve `ErrProtocolLine` wrapping and scanner error behavior.

- [ ] Run `go test ./internal/client -run 'TestRenderTUI' -count=1`.

Expected after refactor: existing TUI renderer tests still pass.

## Task 4: Add TUI Line Forwarder

- [ ] Add test to `internal/client/client_test.go` for a new line-forwarding function.

Required behavior:

- input line `get 旧油灯` is visible in TUI output before submit;
- server receives `get item.tutorial.old_lantern\n`;
- prompt clears after submit.

- [ ] Add function to `internal/client/client.go`:

```go
func ForwardTUILines(input io.Reader, server io.Writer, runtime *TUIRuntime) error
```

Implementation requirements:

- scan input line-by-line;
- call `runtime.SubmitLine(scanner.Text(), server)` for each line;
- return scanner errors and write errors.

- [ ] Run `go test ./internal/client -run 'TestForwardTUILines' -count=1`.

Expected after implementation: pass.

## Task 5: Wire `mudclient --tui` Input Path

- [ ] Modify `cmd/mudclient/main.go` so `--tui` creates one `TUIRuntime` and uses it for both server rendering and stdin line forwarding.

Implementation guidance:

- keep non-TUI mode using `RenderObservedProtocolLines` and `ForwardResolvedCommands`;
- in TUI mode, use `RenderTUIObservedProtocolLines` for server events and `ForwardTUILines` for stdin;
- if sharing one runtime requires a new client-level helper, keep it small and testable.

- [ ] Run `go test ./cmd/mudclient -count=1`.

Expected after wiring: pass.

## Task 6: Verification and Commit

- [ ] Run `gofmt -w internal/client cmd/mudclient`.
- [ ] Run `go test ./internal/client ./cmd/mudclient -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client` and `cmd/mudclient`.
- [ ] Measure pure LOC for changed Go files.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `显示TUI提交输入`.
