# Client Ambiguous Item Feedback Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep ambiguous item names on the client, show candidate feedback locally, and send item commands to the server only when the client resolved a unique item id.

**Architecture:** `State` keeps one shared input-name index for display names and aliases: `input text -> candidate item ids`. A new typed resolution result reports whether a command should be sent and which candidates caused ambiguity. `TUIRuntime` uses that result to suppress ambiguous sends and render a local system event. Non-TUI line forwarding is only a debug/compatibility path; it may keep the safety rule that ambiguous commands are not sent, but feature feedback is TUI-only.

**Tech Stack:** Go stdlib, existing `internal/client`, `internal/client/tui`, `internal/content`, and `internal/protocol` packages.

---

### Task 1: Add Typed Command Resolution

**Files:**
- Modify: `internal/client/state.go`
- Test: `internal/client/state_test.go`

- [ ] **Step 1: Write the failing test**

Add a test that builds two items with the same alias and asserts the new resolution result is ambiguous, not sendable, and exposes both candidates.

```go
func TestState_ResolveCommandInput_reportsAmbiguousCandidates(t *testing.T) {
	// Given
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)

	// When
	got := state.ResolveCommandInput("get shared")

	// Then
	if got.Send {
		t.Fatalf("expected ambiguous command not to send")
	}
	if got.Command != "get shared" {
		t.Fatalf("command = %q, want original command", got.Command)
	}
	wantCandidates := []string{"item.tutorial.old_lantern", "item.tutorial.practice_sword"}
	if !slices.Equal(got.AmbiguousItems, wantCandidates) {
		t.Fatalf("candidates = %#v, want %#v", got.AmbiguousItems, wantCandidates)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/client -run TestState_ResolveCommandInput_reportsAmbiguousCandidates -count=1`

Expected: compile failure because `ResolveCommandInput` and the result type do not exist.

- [ ] **Step 3: Write minimal implementation**

Add:

```go
type CommandResolution struct {
	Command        string
	Send           bool
	AmbiguousItems []string
}
```

Add `ResolveCommandInput(command string) CommandResolution`. It should resolve `get`, `drop`, and `examine`; return `Send: true` for non-item commands, direct ids, unknown names, and unique matches; return `Send: false` with `AmbiguousItems` for ambiguous matches. Keep `ResolveCommand` as a compatibility wrapper returning `.Command`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/client -run TestState_ResolveCommandInput_reportsAmbiguousCandidates -count=1`

Expected: PASS.

### Task 2: Keep Non-TUI Line Input As Safety-Only

**Files:**
- Modify: `internal/client/client.go`
- Test: `internal/client/client_test.go`

- [ ] **Step 1: Write the failing test**

Add a test that calls `ForwardResolvedCommands` with ambiguous input and asserts nothing is sent to the server. Do not add user-facing feedback to this path; ordinary terminal mode is not a target product surface.

```go
func TestForwardResolvedCommands_keepsAmbiguousItemCommandLocal(t *testing.T) {
	// Given
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("get shared\nlook\n")
	var serverOutput strings.Builder

	// When
	err = ForwardResolvedCommands(input, &serverOutput, state)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	want := "look\n"
	if got := serverOutput.String(); got != want {
		t.Fatalf("server output = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/client -run TestForwardResolvedCommands_keepsAmbiguousItemCommandLocal -count=1`

Expected: FAIL because `get shared` is still sent.

- [ ] **Step 3: Write minimal implementation**

In `forwardCommands`, call `state.ResolveCommandInput(line)`. If `Send` is false, skip writing to the server. Keep `ForwardCommands` unchanged when `state == nil`. Do not add local output or rendering to non-TUI mode.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/client -run TestForwardResolvedCommands_keepsAmbiguousItemCommandLocal -count=1`

Expected: PASS.

### Task 3: Show Local TUI Feedback For Ambiguity

**Files:**
- Modify: `internal/client/tui_runtime.go`
- Test: `internal/client/tui_runtime_test.go`

- [ ] **Step 1: Write the failing test**

Add a test that submits an ambiguous item command in TUI mode, asserts nothing reaches the server, and asserts the screen contains the ambiguity text plus candidate display names.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/client -run TestTUIRuntimeSubmitLineShowsAmbiguousItemFeedback -count=1`

Expected: FAIL because the command is still sent and no local feedback appears.

- [ ] **Step 3: Write minimal implementation**

Add a small `State.AmbiguousItemMessage(candidates []string) protocol.Event` helper or equivalent private function in `client` that produces `event=system` with a text key already present in the client catalog if available, or a literal local message if the renderer supports it. Keep this local to the client layer; do not add server behavior.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/client -run TestTUIRuntimeSubmitLineShowsAmbiguousItemFeedback -count=1`

Expected: PASS.

### Task 4: Final Verification

**Files:**
- All modified Go and data files.

- [ ] **Step 1: Format Go files**

Run: `gofmt -w internal/client/state.go internal/client/state_test.go internal/client/client.go internal/client/client_test.go internal/client/tui_runtime.go internal/client/tui_runtime_test.go`

- [ ] **Step 2: Run targeted tests**

Run: `go test ./internal/client -count=1`

Expected: PASS.

- [ ] **Step 3: Run full tests**

Run: `go test -race -shuffle=on -count=1 ./...`

Expected: PASS.

- [ ] **Step 4: Run diagnostics and review LOC by responsibility**

Run diagnostics on changed Go files. Treat 250 lines as a review trigger only when a file mixes responsibilities; cohesive single-behavior files may remain up to 600 lines before requiring stronger decomposition review.
