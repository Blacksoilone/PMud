# Progression/Quest First Version Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a first playable, runtime-only quest system with one tutorial quest, server-side stage progression, and a `quest` command that shows current quest status.

**Architecture:** Keep quest definitions in content, runtime progression state on the server, and player-visible status in presentation/client rendering. The first version is intentionally runtime-only: no persistence, no quest graph branching, and no multi-quest inventory beyond the single active tutorial quest. Session command outcomes feed progression triggers so the quest advances from real player actions instead of client-side guesses.

**Tech Stack:** Go standard library, existing `content`/`world`/`session`/`presentation`/`client` packages, `go test -race -shuffle=on -count=1`, `gofumpt`, `golangci-lint`.

---

## Context

The codebase already has a single authoritative server, a presentation layer, and a text client. There is no quest/progression package yet. The approved quest spec says progression is a stage machine, active quest has exactly one current stage, finish conditions are all required, final stage enters `reward_pending`, and the first version can stay runtime-only.

The first useful quest should reuse existing tutorial content so the feature becomes visible immediately without a new content expansion. The cleanest first quest is a three-step tutorial chain:

1. Get `item.tutorial.old_lantern`.
2. Move to `room.tutorial.yard`.
3. Examine `item.tutorial.practice_sword`.

That gives us a real stage machine, event-driven advancement, and a player-facing status view without inventing persistence or new world mechanics.

Recommended first-slice behavior:
- One active tutorial quest.
- Runtime only; quest state resets with the session.
- A `quest` command shows current quest name, current stage text, and remaining conditions.
- Session emits progression triggers when `get`, `drop`, `move`, and `examine` succeed.
- No branching, no optional objectives, no persistence, no multi-quest log.

## File Map

**Create:**
- `internal/progression/types.go`
- `internal/progression/engine.go`
- `internal/progression/engine_test.go`

**Modify:**
- `internal/content/types.go`
- `internal/content/compiler.go`
- `internal/content/fixture.go`
- `internal/content/compiler_test.go`
- `data/tutorial/source.json`
- `internal/command/command.go`
- `internal/command/command_test.go`
- `internal/session/session.go`
- `internal/session/session_test.go`
- `internal/presentation/event.go`
- `internal/client/render/render.go`
- `internal/client/render/render_test.go`
- `internal/client/client_test.go`
- `internal/client/tui_runtime_test.go`

## Task Dependency Graph

| Task | Depends On | Reason |
|------|------------|--------|
| Task 1 | None | Lock quest content shape and the tutorial quest definition before runtime code. |
| Task 2 | Task 1 | Add the runtime engine and make stage advancement testable in isolation. |
| Task 3 | Task 2 | Wire server commands and action outcomes into the engine and expose a `quest` command. |
| Task 4 | Task 3 | Render quest status in the client and verify the user-visible output. |
| Task 5 | Tasks 1-4 | Final verification and clean up. |

## Parallel Execution Graph

Wave 1:
- Task 1: Quest content shape + tutorial quest definition

Wave 2:
- Task 2: Runtime progression engine

Wave 3:
- Task 3: Session wiring and `quest` command

Wave 4:
- Task 4: Client rendering and tests

Wave 5:
- Task 5: Verification

Critical path: Task 1 → Task 2 → Task 3 → Task 4 → Task 5

---

## Tasks

### Task 1: Add Quest Content Shape And Tutorial Quest Definition

**Description:** Extend content data so quests, stages, conditions, and rewards can be defined alongside rooms and items. Add one tutorial quest to the fixture and tutorial JSON using the existing world content.

**Files:**
- Modify: `internal/content/types.go`
- Modify: `internal/content/compiler.go`
- Modify: `internal/content/fixture.go`
- Modify: `internal/content/compiler_test.go`
- Modify: `data/tutorial/source.json`

**Acceptance Criteria:**
- The content source can describe a quest with exactly one current stage at runtime.
- The tutorial quest has three stages:
  - stage 1 completes when the player gets `item.tutorial.old_lantern`
  - stage 2 completes when the player moves to `room.tutorial.yard`
  - stage 3 completes when the player examines `item.tutorial.practice_sword`
- Quest text keys and stage text keys compile into the client catalog the same way room/item text already does.
- `internal/content/compiler_test.go` proves the quest data survives compile.

**Test First:**

```go
func TestCompile_projectsTutorialQuest(t *testing.T) {
	compiled, err := Compile(TutorialSource())
	if err != nil {
		t.Fatal(err)
	}

	quest, ok := compiled.Server.Quests["quest.tutorial.first_steps"]
	if !ok {
		t.Fatal("missing tutorial quest")
	}
	if quest.StageIDs[0] != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage 0 = %q", quest.StageIDs[0])
	}
	if got := compiled.Client.Text[quest.NameKey]; got != "教程任务" {
		t.Fatalf("quest name = %q", got)
	}
}
```

**TDD Steps:**
- [ ] **Step 1: Write the failing test**
- [ ] **Step 2: Run it**
  - Run: `go test ./internal/content -run TestCompile_projectsTutorialQuest -count=1`
  - Expected: fail because quest types and compiled maps do not exist yet.
- [ ] **Step 3: Write minimal content model support**
  - Add quest source, stage source, condition source, and reward source types.
  - Add compiled quest types to `ServerSnapshot` and `ClientCatalog`.
  - Populate the tutorial quest in `fixture.go` and `data/tutorial/source.json`.
- [ ] **Step 4: Run the test and make it pass**
  - Run: `go test ./internal/content -count=1`
- [ ] **Step 5: Commit**
  - Commit after content data compiles cleanly.

---

### Task 2: Add Runtime Progression Engine

**Description:** Create a new server-side progression package that owns the active quest state, stage advancement, and condition evaluation for one player session.

**Files:**
- Create: `internal/progression/types.go`
- Create: `internal/progression/engine.go`
- Create: `internal/progression/engine_test.go`

**Recommended API:**

```go
type TriggerKind string

const (
	TriggerGotItem      TriggerKind = "got_item"
	TriggerMovedRoom    TriggerKind = "moved_room"
	TriggerExaminedItem TriggerKind = "examined_item"
)

type Trigger struct {
	Kind TriggerKind
	ItemID string
	RoomID string
}

type Status struct {
	QuestID     string
	QuestName   string
	StageID     string
	StageText   string
	Conditions  []string
	State       string
}

type Engine struct { ... }

func NewEngine(defs Definitions) *Engine
func (e *Engine) Apply(playerID string, trigger Trigger) (Status, bool)
func (e *Engine) Status(playerID string) (Status, bool)
```

**Acceptance Criteria:**
- Exactly one active quest exists for the first version.
- Stage completion requires all listed conditions to be satisfied.
- Quest state transitions through `active` → `reward_pending` → `completed`.
- The engine is runtime-only; no persistence or save/load support.
- The engine can answer current status at any time for a `quest` command.

**Test First:**

```go
func TestEngine_advancesTutorialQuestStages(t *testing.T) {
	engine := NewEngine(TutorialDefinitions())
	playerID := "player.local"

	status, advanced := engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})
	if !advanced || status.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("status after get = %#v, advanced=%v", status, advanced)
	}

	status, advanced = engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})
	if !advanced || status.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("status after move = %#v, advanced=%v", status, advanced)
	}

	status, advanced = engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})
	if !advanced || status.State != "reward_pending" {
		t.Fatalf("status after examine = %#v, advanced=%v", status, advanced)
	}
}
```

**TDD Steps:**
- [ ] **Step 1: Write the failing test**
- [ ] **Step 2: Run it**
  - Run: `go test ./internal/progression -run TestEngine_advancesTutorialQuestStages -count=1`
  - Expected: fail because the package and methods do not exist yet.
- [ ] **Step 3: Implement the minimal engine**
  - Add stage-machine state and trigger matching.
  - Keep the first quest definition in code or injected definitions for now.
- [ ] **Step 4: Run the test and make it pass**
  - Run: `go test ./internal/progression -count=1`
- [ ] **Step 5: Commit**
  - Commit the progression engine once it is green.

---

### Task 3: Wire Session Commands And Quest Querying

**Description:** Make the server feed quest triggers from real command outcomes and add a `quest` command so the player can inspect current quest state.

**Files:**
- Modify: `internal/command/command.go`
- Modify: `internal/command/command_test.go`
- Modify: `internal/session/session.go`
- Modify: `internal/session/session_test.go`
- Modify: `internal/presentation/event.go`

**Recommended Event Type:**

```go
type QuestStatusEvent struct {
	QuestID    string
	QuestName  string
	StageID    string
	StageText  string
	Conditions []string
	State      string
}
```

**Acceptance Criteria:**
- `ParseClientInput("quest")` and `ParseServerInput("quest")` yield a new `QuestCommand`.
- `sessionState.handleLine("quest")` returns the current quest status.
- Successful `get`, `move`, and `examine` outcomes feed the progression engine:
  - `system.item.taken` for the lantern advances stage 1
  - room change to yard advances stage 2
  - `ItemObservationEvent` for the practice sword advances stage 3
- Quest state stays server-owned and runtime-only.
- No persistence or save file support is added.

**Test First:**

```go
func TestSessionQuest_reportsTutorialStatus(t *testing.T) {
	state := newTestSessionState()
	event := state.handleLine("quest")
	quest, ok := event.(presentation.QuestStatusEvent)
	if !ok {
		t.Fatalf("expected quest status, got %T", event)
	}
	if quest.QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("quest id = %q", quest.QuestID)
	}
}
```

```go
func TestSessionGet_advancesQuestWhenLanternTaken(t *testing.T) {
	state := newTestSessionState()
	state.handleLine("get 旧油灯")

	event := state.handleLine("quest")
	quest := event.(presentation.QuestStatusEvent)
	if quest.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("stage id = %q", quest.StageID)
	}
}
```

**TDD Steps:**
- [ ] **Step 1: Write the failing tests**
- [ ] **Step 2: Run them**
  - Run: `go test ./internal/command ./internal/session -run 'Test.*Quest|TestSession.*Tutorial' -count=1`
  - Expected: fail because `QuestCommand`/`QuestStatusEvent` and quest wiring do not exist yet.
- [ ] **Step 3: Wire the engine**
  - Add the `QuestCommand` to parser and server dispatch.
  - Add a progression engine instance to `sessionState`.
  - Feed triggers after successful `get`, `move`, and `examine` outcomes.
- [ ] **Step 4: Run the tests and make them pass**
  - Run: `go test ./internal/command ./internal/session -count=1`
- [ ] **Step 5: Commit**
  - Save the session wiring once green.

---

### Task 4: Render Quest Status In The Client

**Description:** Teach the client renderer to show the quest status event so the first quest is visible in the TUI and text client.

**Files:**
- Modify: `internal/client/render/render.go`
- Modify: `internal/client/render/render_test.go`
- Modify: `internal/client/client_test.go`
- Modify: `internal/client/tui_runtime_test.go`

**Acceptance Criteria:**
- `presentation.QuestStatusEvent` renders into a human-readable block showing quest name, stage text, current state, and remaining conditions.
- The client TUI can display the `quest` command output.
- Quest rendering does not disturb existing room, inventory, or item output.

**Suggested Render Shape:**

```text
任务: 教程第一步
阶段: 拿起旧油灯
状态: active
条件:
- 获取旧油灯
```

**Test First:**

```go
func TestRenderQuestStatusEvent(t *testing.T) {
	got := Render(protocol.Event{
		Name: "QuestStatusEvent",
		Fields: map[string]string{
			"quest_id": "quest.tutorial.first_steps",
			"quest_name": "教程第一步",
			"stage_text": "拿起旧油灯",
			"state": "active",
			"conditions": "获取旧油灯",
		},
	})
	want := "任务: 教程第一步\n阶段: 拿起旧油灯\n状态: active\n条件:\n- 获取旧油灯\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
```

**TDD Steps:**
- [ ] **Step 1: Write the failing render test**
- [ ] **Step 2: Run it**
  - Run: `go test ./internal/client/render -run TestRenderQuestStatusEvent -count=1`
  - Expected: fail because the renderer does not know quest events yet.
- [ ] **Step 3: Implement the renderer case**
  - Add a quest event branch in `render.go`.
  - Keep existing render behavior unchanged.
- [ ] **Step 4: Run client tests and make them pass**
  - Run: `go test ./internal/client/render ./internal/client -count=1`
- [ ] **Step 5: Commit**
  - Save the rendering update once green.

---

### Task 5: Final Verification And Save

**Description:** Validate the whole first-version quest slice end-to-end, then save progress.

**Files:**
- None

**Acceptance Criteria:**
- All changed Go files have clean diagnostics.
- Targeted package tests pass.
- Full race/shuffle suite passes.
- No persistence or multi-quest scope has slipped in.

**Commands:**

```bash
gofumpt -l .
golangci-lint run ./...
go test ./internal/progression ./internal/content ./internal/command ./internal/session ./internal/world ./internal/client -count=1
go test -race -shuffle=on -count=1 ./...
```

**Stop Criteria:**
- Stop if the implementation starts requiring save/load, multi-quest logs, branching quests, or a generalized task engine.
- Stop if the first tutorial quest cannot be expressed with the existing room/item/action model and a single runtime engine.

## Missing Gaps Check

Spec coverage check:
- Stage machine: covered by Task 2.
- Single current stage: covered by Task 2.
- Finish conditions all required: covered by Task 2.
- Reward pending before completion: covered by Task 2.
- Runtime-only first version: covered in the architecture and Task 2.
- UI exposure: covered by Task 4.
- Event-driven advancement: covered by Task 3.

No blockers remain beyond the runtime-only assumption, which is intentionally locked for this first version.
