# Data Driven Tutorial World Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Load the tutorial world from JSON data instead of constructing production content through `content.TutorialSource()`.

**Architecture:** Add a JSON boundary loader in `internal/content`, store the current tutorial source in `data/tutorial/source.json`, and switch `mudserver`/`mudclient` startup to compile that data file. Keep `TutorialSource()` as a test fixture so existing tests stay concise.

**Tech Stack:** Go standard library `encoding/json`, existing `internal/content`, `internal/world`, `cmd/mudserver`, and `cmd/mudclient` packages.

---

## File Structure

- Create `internal/content/json_loader.go`: load a `ContentSource` from a JSON file path.
- Create `internal/content/json_loader_test.go`: loader tests for valid JSON, malformed JSON, and missing files.
- Create `data/tutorial/source.json`: data copy of the current tutorial fixture.
- Modify `cmd/mudserver/main.go`: compile world from `data/tutorial/source.json`.
- Add/modify `cmd/mudserver/main_test.go`: prove `buildGame` can load the JSON data file.
- Modify `cmd/mudclient/main.go`: compile `ClientCatalog` from `data/tutorial/source.json`.
- Modify `cmd/mudclient/main_test.go`: prove the client catalog loader reads the JSON data file.

## Task 1: JSON Loader Tests

- [ ] Add `internal/content/json_loader_test.go` first.

Required tests:

- valid JSON loads into `ContentSource` and can be compiled;
- malformed JSON returns an error;
- missing file returns an error.

Run:

```bash
go test ./internal/content -run 'TestLoadSource' -count=1
```

Expected before implementation: compile failure for missing `LoadSource`.

## Task 2: Implement JSON Loader

- [ ] Add `internal/content/json_loader.go`.

Required shape:

```go
func LoadSource(path string) (ContentSource, error)
```

Implementation requirements:

- read the file with `os.ReadFile`;
- parse with `encoding/json` into `ContentSource`;
- wrap read and parse errors with context using `%w`;
- do not validate gameplay references here; `Compile` remains the compiler boundary.

Run:

```bash
go test ./internal/content -run 'TestLoadSource' -count=1
```

Expected after implementation: pass.

## Task 3: Tutorial JSON Data

- [ ] Add `data/tutorial/source.json` with the same rooms, items, and text as `content.TutorialSource()`.
- [ ] Add a test that loads `data/tutorial/source.json`, compiles it, and compares the compiled server/client content to `content.TutorialSource()` compiled output.

Run:

```bash
go test ./internal/content -run 'TestLoadTutorialSourceJSONMatchesFixture' -count=1
```

Expected after data file is added: pass.

## Task 4: Server Startup Uses Data File

- [ ] Modify `cmd/mudserver/main.go`.

Required behavior:

- `main` loads `data/tutorial/source.json` by default;
- `buildGame(path string)` returns `(*world.World, error)`;
- `main` panics or prints a fatal error if loading/compiling fails;
- production startup no longer calls `content.TutorialSource()`.

- [ ] Add `cmd/mudserver/main_test.go` to call `buildGame("../../data/tutorial/source.json")`.

Run:

```bash
go test ./cmd/mudserver -count=1
```

Expected after implementation: pass.

## Task 5: Client Startup Uses Data File

- [ ] Modify `cmd/mudclient/main.go`.

Required behavior:

- client startup loads `data/tutorial/source.json` by default;
- extract a small `loadClientCatalog(path string)` helper returning `content.ClientCatalog, error`;
- production startup no longer calls `content.TutorialSource()`.

- [ ] Add tests in `cmd/mudclient/main_test.go` for `loadClientCatalog("../../data/tutorial/source.json")`.

Run:

```bash
go test ./cmd/mudclient -count=1
```

Expected after implementation: pass.

## Task 6: Verification and Commit

- [ ] Run `gofmt -w internal/content cmd/mudserver cmd/mudclient`.
- [ ] Run `go test ./internal/content ./cmd/mudserver ./cmd/mudclient -count=1`.
- [ ] Run `go test -race -shuffle=on -count=1 ./...`.
- [ ] Run LSP diagnostics on changed Go packages.
- [ ] Inspect `GIT_MASTER=1 git status --short`, targeted diff, and `GIT_MASTER=1 git log --oneline -10`.
- [ ] Stage focused files and commit with `从数据加载教程世界`.
