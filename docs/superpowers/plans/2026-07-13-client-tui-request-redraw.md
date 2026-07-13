# Client TUI Request Redraw Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `mudclient --tui` refresh the same terminal area on each server event instead of appending a new boxed screen below the previous one.

**Architecture:** Add a small `internal/client/screen` package that owns ANSI full-screen redraw. Keep refresh request-driven: the existing TUI protocol renderer calls `Draw` only after receiving and applying a server event. Do not introduce dirty diffing, a frame ticker, raw mode, or cursor lifecycle management in this slice.

**Tech Stack:** Go standard library, existing `internal/client/layout`, `internal/client/tui`, and `internal/client` packages.

---

## File Structure

- Create `internal/client/screen/screen.go`: full redraw renderer.
- Create `internal/client/screen/screen_test.go`: ANSI redraw behavior tests.
- Modify `internal/client/client.go`: route TUI protocol output through `screen.Renderer`.
- Modify `internal/client/client_test.go`: assert TUI renderer emits redraw control sequences per event.

## Task 1: Screen Renderer Tests

- [ ] Add `internal/client/screen/screen_test.go` first.

Required tests:

```go
func TestRendererDrawWritesFullRedraw(t *testing.T) {
	var output strings.Builder
	renderer := screen.NewRenderer(&output)
	block := layout.NewBlock([]string{"hello", "world"})

	err := renderer.Draw(block)

	if err != nil {
		t.Fatalf("Draw: %v", err)
	}
	want := "\x1b[2J\x1b[Hhello\nworld\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}
```

```go
func TestRendererDrawPropagatesWriteError(t *testing.T) {
	renderer := screen.NewRenderer(failingWriter{})
	block := layout.NewBlock([]string{"hello"})

	err := renderer.Draw(block)

	if err == nil {
		t.Fatalf("Draw error = nil, want write error")
	}
}
```

- [ ] Run `go test ./internal/client/screen -count=1`.

Expected before implementation: package or symbol compile failure.

## Task 2: Implement Screen Renderer

- [ ] Create `internal/client/screen/screen.go`.

Required shape:

```go
package screen

import (
	"PMud/internal/client/layout"
	"io"
)

const FullRedrawPrefix = "\x1b[2J\x1b[H"

type Renderer struct {
	output io.Writer
}

func NewRenderer(output io.Writer) Renderer
func (r Renderer) Draw(block layout.Block) error
```

Implementation requirements:

- `Draw` writes `FullRedrawPrefix` first;
- `Draw` then writes `block.String()`;
- return write errors immediately;
- do not hide/show cursor yet.

- [ ] Run `go test ./internal/client/screen -count=1`.

Expected after implementation: pass.

## Task 3: Client TUI Redraw Integration Test

- [ ] Add a test to `internal/client/client_test.go`.

Required behavior:

- `RenderTUIObservedProtocolLines` emits `screen.FullRedrawPrefix` once per parsed server event;
- rendered content still includes the prompt and event text.

Suggested test body:

```go
func TestRenderTUIObservedProtocolLines_redrawsPerEvent(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader(
		"event=system\tmessage_key=system.help\n" +
			"event=inventory\titems=item.tutorial.old_lantern\n",
	)
	var output strings.Builder

	err = RenderTUIObservedProtocolLines(input, &output, state, 48, 3)

	if err != nil {
		t.Fatalf("RenderTUIObservedProtocolLines: %v", err)
	}
	got := output.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "可用命令") || !strings.Contains(got, "你带着: 旧油灯") || !strings.Contains(got, "> ") {
		t.Fatalf("output missing expected TUI content:\n%s", got)
	}
}
```

- [ ] Run `go test ./internal/client -run 'TestRenderTUIObservedProtocolLines_redrawsPerEvent' -count=1`.

Expected before integration: test fails because no redraw prefix is emitted.

## Task 4: Wire Screen Renderer Into TUI Protocol Renderer

- [ ] Modify `internal/client/client.go`.

Implementation requirements:

- import `PMud/internal/client/screen`;
- create one `screen.Renderer` before scanning input;
- after each event, call `renderer.Draw(tui.View(model, state.catalog, width))`;
- keep parse error and scanner error behavior unchanged.

- [ ] Run `go test ./internal/client -run 'TestRenderTUI' -count=1`.

Expected after implementation: pass.

## Task 5: Verification and Commit

- [ ] Run `gofmt -w internal/client/screen internal/client/client.go internal/client/client_test.go`.
- [ ] Run `go test ./internal/client/screen ./internal/client -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client/screen` and `internal/client`.
- [ ] Measure pure LOC for changed Go files.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage only this slice and commit with `刷新TUI时重绘屏幕`.
