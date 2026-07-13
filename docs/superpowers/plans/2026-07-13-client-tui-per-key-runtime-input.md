# Client TUI Per-Key Runtime Input Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let `TUIRuntime` handle individual `tui.Input` actions so the prompt redraws immediately for text, backspace, clear, and submit operations.

**Architecture:** Keep raw terminal byte parsing out of this slice. The runtime exposes a typed input-action method that mutates the existing `tui.Model`, redraws after each visible change, and writes a resolved command only on submit. Existing line-based input should reuse this path so future raw-key input and current stdin line input share behavior.

**Tech Stack:** Go standard library, existing `internal/client`, `internal/client/tui`, `internal/client/screen`, and `internal/protocol` packages.

---

## File Structure

- Modify `internal/client/tui_runtime.go`: add `ApplyInput(input tui.Input, server io.Writer) error` and refactor `SubmitLine` through it.
- Modify `internal/client/tui_runtime_test.go`: add per-key text/backspace/clear/submit tests.

## Task 1: Per-Key Text Redraw Test

- [ ] Add test to `internal/client/tui_runtime_test.go`:

```go
func TestTUIRuntimeApplyInputRedrawsText(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "get "}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "旧油灯"}, &server)

	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	got := output.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("output missing typed prompt:\n%s", got)
	}
	if server.String() != "" {
		t.Fatalf("server output = %q, want empty", server.String())
	}
}
```

- [ ] Run `go test ./internal/client -run 'TestTUIRuntimeApplyInputRedrawsText' -count=1`.

Expected before implementation: compile failure for missing `ApplyInput` method.

## Task 2: Implement Text Input Redraw

- [ ] Add method to `internal/client/tui_runtime.go`:

```go
func (r *TUIRuntime) ApplyInput(input tui.Input, server io.Writer) error
```

Implementation requirements:

- lock the runtime mutex;
- apply non-submit input with `tui.ApplyInput`;
- redraw once;
- do not write to server for text/backspace/clear.

- [ ] Run `go test ./internal/client -run 'TestTUIRuntimeApplyInputRedrawsText' -count=1`.

Expected after implementation: pass.

## Task 3: Backspace and Clear Tests

- [ ] Add tests to `internal/client/tui_runtime_test.go`:

```go
func TestTUIRuntimeApplyInputBackspaceRemovesLastRune(t *testing.T) { ... }
func TestTUIRuntimeApplyInputClearClearsPrompt(t *testing.T) { ... }
```

Required behavior:

- after typing `get 旧油灯` and backspace, the final output contains `> get 旧油` and not `> get 旧油灯` in the last frame;
- after clear, the final output contains cleared prompt `| >`.

- [ ] Run `go test ./internal/client -run 'TestTUIRuntimeApplyInputBackspace|TestTUIRuntimeApplyInputClear' -count=1`.

Expected before implementation if text-only code exists: backspace/clear may already pass through generic non-submit handling. If they pass immediately, keep them as regression coverage because the behavior already exists through `tui.ApplyInput`.

## Task 4: Submit Action Test

- [ ] Add test to `internal/client/tui_runtime_test.go`:

```go
func TestTUIRuntimeApplyInputSubmitWritesResolvedCommandAndClearsPrompt(t *testing.T) { ... }
```

Required behavior:

- observe old lantern in state first;
- apply text `get 旧油灯`;
- apply submit;
- server receives `get item.tutorial.old_lantern\n`;
- output includes the typed prompt and a later cleared prompt.

- [ ] Run `go test ./internal/client -run 'TestTUIRuntimeApplyInputSubmit' -count=1`.

Expected before implementation: fail because submit does not write server or clear through the new method.

## Task 5: Implement Submit and Refactor SubmitLine

- [ ] Extend `TUIRuntime.ApplyInput`:

Implementation requirements:

- for `tui.InputSubmit`, apply input, resolve non-empty command through `State.ResolveCommand`, write resolved line plus newline to `server`, and redraw cleared prompt;
- return write errors;
- `SubmitLine(line, server)` should call `ApplyInput(InputText{line})` then `ApplyInput(InputSubmit)` so line input and future raw input share behavior.

- [ ] Run `go test ./internal/client -run 'TestTUIRuntimeApplyInput|TestTUIRuntimeSubmitLine|TestForwardTUILines' -count=1`.

Expected after implementation: pass.

## Task 6: Verification and Commit

- [ ] Run `gofmt -w internal/client/tui_runtime.go internal/client/tui_runtime_test.go`.
- [ ] Run `go test ./internal/client -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on `internal/client`.
- [ ] Measure pure LOC for changed Go files.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `支持TUI逐键输入动作`.
