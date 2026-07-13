# Client TUI Runtime Skeleton Implementation Plan

> Required sub-skills: `test-driven-development` and `programming` before coding Go files.

## Goal

Add a coherent `internal/client/tui` runtime-skeleton package that owns terminal-facing client state, input-buffer updates, command submission, and a plain view rendered with the existing layout/render primitives.

## Architecture

The TUI layer stays pure and testable. It stores `protocol.Event` values, not rendered strings, so catalog lookup and layout happen in `View`. It does not parse server protocol lines, resolve item display names, write to sockets, or choose final ASCII art direction.

## Tech Stack

Go standard library, existing `internal/client/layout`, `internal/client/render`, `internal/content`, and `internal/protocol` packages.

---

## File Structure

- `internal/client/tui/model.go`: `Model`, `Command`, constructors, event history update.
- `internal/client/tui/input.go`: input action types and `ApplyInput`.
- `internal/client/tui/view.go`: `View` composition using render/layout.
- `internal/client/tui/model_test.go`: model and history tests.
- `internal/client/tui/input_test.go`: input editing and submit tests.
- `internal/client/tui/view_test.go`: rendered view tests with catalog text and CJK prompt.

## Task 1: Add Model and Event History Tests

Create `internal/client/tui/model_test.go` first.

Test cases:

- `NewModel(2)` starts with empty events and empty input.
- `ApplyEvent` appends protocol events.
- when the history limit is exceeded, the oldest event is dropped.
- `NewModel(0)` should still keep at least one event, so callers cannot create a useless model accidentally.

Expected command before implementation:

```bash
go test ./internal/client/tui -run 'TestModel|TestApplyEvent' -count=1
```

Expected initial output: package or symbol compile failure.

## Task 2: Implement Model and ApplyEvent

Create `internal/client/tui/model.go`.

Required shape:

```go
package tui

import "PMud/internal/protocol"

type Model struct {
    Events       []protocol.Event
    Input        string
    HistoryLimit int
}

type Command struct {
    Line string
}

func NewModel(historyLimit int) Model
func ApplyEvent(model Model, event protocol.Event) Model
```

Implementation requirements:

- value-style updates: return a new `Model` value;
- copy event slices before appending so callers cannot mutate prior model history through shared backing arrays;
- clamp `historyLimit < 1` to `1`;
- keep only the newest `HistoryLimit` events.

Run:

```bash
go test ./internal/client/tui -run 'TestModel|TestApplyEvent' -count=1
```

Expected output: pass.

## Task 3: Add Input Update Tests

Create `internal/client/tui/input_test.go`.

Test cases:

- text input appends to the current buffer;
- backspace removes the last rune, not the last byte;
- clear empties the buffer;
- submit returns `Command{Line: input}` and clears the buffer;
- submitting whitespace-only input returns an empty command and clears the buffer.

Expected command before implementation:

```bash
go test ./internal/client/tui -run 'TestApplyInput' -count=1
```

Expected initial output: missing input symbols.

## Task 4: Implement Input Actions

Create `internal/client/tui/input.go`.

Required shape:

```go
type InputKind int

const (
    InputText InputKind = iota
    InputBackspace
    InputClear
    InputSubmit
)

type Input struct {
    Kind InputKind
    Text string
}

func ApplyInput(model Model, input Input) (Model, Command)
```

Implementation requirements:

- `InputText` appends `input.Text` exactly;
- `InputBackspace` handles UTF-8 runes correctly;
- `InputSubmit` trims only for the emptiness check, but returns the original non-empty input line;
- no command parsing or item-name resolution in this package.

Run:

```bash
go test ./internal/client/tui -run 'TestApplyInput' -count=1
```

Expected output: pass.

## Task 5: Add View Tests

Create `internal/client/tui/view_test.go`.

Test cases:

- view with a room event includes catalog room name, description, item name, and prompt line;
- view with multiple events includes both rendered events in order;
- view with CJK input keeps the prompt text visible after boxing/composition.

Use existing `content.TutorialSource()` and `content.Compile(...)` fixtures when possible to avoid inventing another catalog shape.

Expected command before implementation:

```bash
go test ./internal/client/tui -run 'TestView' -count=1
```

Expected initial output: missing `View` symbol.

## Task 6: Implement View

Create `internal/client/tui/view.go`.

Required shape:

```go
func View(model Model, catalog content.ClientCatalog, width int) layout.Block
```

Implementation requirements:

- render each stored event with `render.RenderBlock`;
- compose event blocks vertically;
- render prompt as a `layout.Block` containing `> ` plus current input;
- use `layout.Box` around the event area and input area so this becomes a true runtime skeleton, not another render wrapper;
- keep the implementation plain and deterministic.

Run:

```bash
go test ./internal/client/tui -run 'TestView' -count=1
```

Expected output: pass.

## Task 7: Verify the Package and Workspace

Run:

```bash
gofmt -w internal/client/tui
go test ./internal/client/tui -count=1
go test -race -shuffle=on -count=1 ./...
```

Then run LSP diagnostics on changed Go files:

- `internal/client/tui/model.go`
- `internal/client/tui/input.go`
- `internal/client/tui/view.go`
- all new `_test.go` files

Expected output: tests pass and diagnostics report no errors.

## Task 8: Commit

Before committing, inspect:

```bash
git status --short
git diff -- docs/client-tui-runtime-skeleton.md docs/superpowers/plans/2026-07-13-client-tui-runtime-skeleton.md internal/client/tui
git log --oneline -10
```

Stage only the design doc, plan, and `internal/client/tui` files.

Suggested commit message:

```text
新增客户端TUI运行骨架
```

Expected result: one focused commit containing the grouped TUI runtime-skeleton slice.
