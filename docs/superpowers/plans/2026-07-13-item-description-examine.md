# Item Description Examine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let data-defined items carry description text and expose it through an `examine <item>` command.

**Architecture:** Extend the content compiler first, then world visibility lookup, then presentation/client rendering, then session command handling. `examine` accepts item IDs on the server side, matching the existing `get/drop` contract; client-side display-name resolution can map `examine 旧油灯` to `examine item.tutorial.old_lantern` through the existing state path.

**Tech Stack:** Go standard library, existing `internal/content`, `internal/world`, `internal/presentation`, `internal/client/render`, `internal/session`, and JSON tutorial data.

---

## File Structure

- Modify `internal/content/types.go`: add item description keys to source/server/client structures.
- Modify `internal/content/compiler.go`: compile item description keys.
- Modify `internal/content/fixture.go` and `data/tutorial/source.json`: add tutorial item descriptions.
- Modify `internal/content/json_loader_test.go` and/or compiler tests: verify JSON and fixture compile equivalently with item descriptions.
- Modify `internal/world/types.go` and `internal/world/items.go`: expose item description lookup when visible in room or inventory.
- Modify `internal/world/world_test.go`: test examine visibility.
- Modify `internal/presentation/event.go`, `text_renderer.go`, and tests: add item observation event.
- Modify `internal/client/render/render.go` and tests: render examined item description through catalog.
- Modify `internal/session/session.go` and tests: add `examine <item>` command.

## Task 1: Content Item Description Key

- [ ] Add a failing content test asserting `ItemSource.DescriptionKey` compiles into both server and client content.
- [ ] Run `go test ./internal/content -run 'TestCompile.*ItemDescription|TestLoadTutorialSourceJSONMatchesFixture' -count=1`; expected failure: missing `DescriptionKey` fields.
- [ ] Add `DescriptionKey content.TextKey` to item source/server/client structures and compile it.
- [ ] Add description keys/text to `TutorialSource()` and `data/tutorial/source.json`.
- [ ] Run the same tests; expected pass.

## Task 2: World Examine Lookup

- [ ] Add world tests for examining an item in the current room, in inventory, and not visible.
- [ ] Run `go test ./internal/world -run 'TestWorld_ExamineItem' -count=1`; expected failure: missing `ExamineItem`.
- [ ] Implement `World.ExamineItem(roomID, itemID, playerID)` returning item ID, name key/name, description key/description, and visibility bool.
- [ ] Run the world tests; expected pass.

## Task 3: Protocol and Client Rendering

- [ ] Add presentation test for an `ItemObservationEvent` serialized as `event=item	item=...	name_key=...	description_key=...`.
- [ ] Add client render test for `event=item`, expecting name and description text from the catalog.
- [ ] Run `go test ./internal/presentation ./internal/client/render -run 'Test.*Item' -count=1`; expected failure: missing event/render support.
- [ ] Implement the presentation event and client render branch.
- [ ] Run the same tests; expected pass.

## Task 4: Session Examine Command

- [ ] Add session tests for `examine item.tutorial.old_lantern` in room, after pickup from inventory, and invisible item returning `system.item.not_here`.
- [ ] Run `go test ./internal/session -run 'TestSessionExamine' -count=1`; expected failure: command returns unknown.
- [ ] Implement `examine <item>` in `sessionState.handleLine` using world examine lookup and item observation event.
- [ ] Run the session tests; expected pass.

## Task 5: Verification and Commit

- [ ] Run `gofmt -w internal/content internal/world internal/presentation internal/client/render internal/session`.
- [ ] Run `go test ./internal/content ./internal/world ./internal/presentation ./internal/client/render ./internal/session -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on changed packages.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `支持查看物品描述`.
