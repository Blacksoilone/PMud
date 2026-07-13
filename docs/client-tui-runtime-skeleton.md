# Client TUI Runtime Skeleton

## Goal

Build the next client slice as one coherent runtime-skeleton group, not as isolated rendering wrappers. The slice should make the terminal client ready for a real TUI without committing to the final visual design.

This is still bottom-layer work: it prepares state, update, and view primitives that later art direction and interaction design can use.

## Scope

Add a small `internal/client/tui` package with pure, testable components:

- `Model`: keeps recent server events, the current input line, and a history limit.
- `Update`: applies server events and local input actions to the model.
- `View`: renders the model into the existing `layout.Block` system.

The first view is intentionally plain:

- recent events appear in an output area using `render.RenderBlock`;
- the current input appears as a prompt line such as `> get 旧油灯`;
- composition uses existing `layout.VBox`, `layout.Box`, and related primitives.

## Non-Goals

- Do not design the final ASCII-art screen.
- Do not introduce a TUI framework in this slice.
- Do not add a new wrapper package around `render.RenderBlock` unless the runtime model actually needs it.
- Do not move server authority or command parsing into the TUI package.

## Package Shape

`internal/client/tui` owns terminal-facing client state only.

Suggested public surface:

```go
type Model struct { ... }

func NewModel(historyLimit int) Model
func ApplyEvent(model Model, event protocol.Event) Model
func ApplyInput(model Model, input Input) (Model, Command)
func View(model Model, catalog content.ClientCatalog, width int) layout.Block
```

`Input` should model simple local editing actions: append text, backspace, submit, and clear. `Command` should be empty unless submit produces a line to send to the existing client command pipeline.

## Data Flow

Server protocol events stay as `protocol.Event` values. The TUI model stores them as events, not pre-rendered text, so catalog lookup and view layout remain deterministic at render time.

Local input actions update only the input buffer. Submit returns the current command and clears the buffer. The caller remains responsible for resolving display names and writing to the server.

## Testing

Tests should cover the grouped behavior:

- events append to history;
- history limit drops oldest events;
- input editing changes the buffer predictably;
- submit returns a command and clears the buffer;
- view output includes rendered event text and the prompt line;
- CJK text remains aligned through the existing layout primitives.

## First Integration

After the pure package is covered, connect it lightly to `cmd/mudclient` only if the resulting loop can remain small. If terminal raw-mode handling would dominate the slice, keep integration as a follow-up and leave this slice as a verified runtime model plus view foundation.
