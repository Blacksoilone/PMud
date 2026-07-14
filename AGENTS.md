# AGENTS.md

This file is the primary project guide for coding agents.

## Authority

1. The user's latest direct instruction wins.
2. This root `AGENTS.md` is the default project authority.
3. Current code and tests are the next source of truth.
4. `docs/` is historical unless this file explicitly lists a document as authoritative.
5. If an authoritative doc conflicts with this file, this file wins.

Do not mine unlisted `docs/` files for requirements without asking the user.

Authoritative supplementary specs:

- `docs/superpowers/specs/2026-07-13-tag-definition-instance-design.md`
- `docs/superpowers/specs/2026-07-14-simplified-progression-quest-system.md`

## Project Shape

- This is a Go MUD project.
- The server is authoritative for world state, rules, and actions.
- The TUI is the primary product surface.
- Ordinary stdin/terminal mode is debug/compatibility only. Do not expand its UX unless the user asks.

## Development Workflow

- Use TDD for behavior changes: write the failing behavior test first, then implement.
- Test behavior boundaries, not every file or private helper.
- Prefer useful vertical slices. Do not split progress into tiny Zeno-style tasks.
- After edits, run diagnostics on changed Go files.
- Run targeted tests first, then:

```sh
go test -race -shuffle=on -count=1 ./...
```

- Never delete or weaken failing tests to pass.

## File Size And Refactoring

- Behavior cohesion is primary; line count is secondary.
- `<=250` lines is not a problem.
- `>250` lines is only a review trigger if the file mixes behaviors or is hard to navigate.
- A cohesive single-behavior file may remain intact until about `600` lines.
- Above `600` lines, stop and consider helpers, branch reduction, or registry/dispatcher design.
- Do not split files just to satisfy a numeric limit.

## Package Boundaries

- `internal/command`: command parsing and canonicalization.
- `internal/client`: TUI/client state, item display-name and alias resolution, ambiguity feedback, local static commands.
- `internal/session`: server command dispatch from canonical commands to world actions.
- `internal/world`: authoritative world graph, state, and rules. No text command parsing.
- `internal/content`: content source, compiler, server snapshot, client catalog projection.
- `internal/presentation`: server event to wire protocol.
- `internal/client/render`: wire protocol to player-facing text.

## Command System

- TUI input goes through `internal/command.ParseClientInput`.
- The client resolves item names/aliases to item ids.
- Ambiguous item names are handled locally in the TUI and are not sent to the server.
- The client sends canonical wire commands.
- The server parses canonical wire commands with `internal/command.ParseServerInput`.
- The server must not parse item display names or client item aliases.

Client-only aliases include:

- `take <item>` -> `get <item>`
- `x <item>` / `inspect <item>` -> `examine <item>`
- `i` -> `inventory`
- `l` -> `look`
- bare standard directions -> `go <direction>`

Standard direction aliases include:

- `n/s/e/w/u/d`
- `ne/nw/se/sw`
- full names such as `northwest`
- existing Chinese `北/南`

Do not create implicit aliases for special exits. Special exit aliases must come from content data if they are ever supported.

`help` and empty input are client-local static TUI responses. The server may keep defensive handling, but it is not the product path.

## Data And Code Boundary

- Content data must never contain executable code.
- Even one-off behavior must live in code and be invoked by data.
- Data declares ids, parameters, tags, conditions, and rewards.
- Code implements behavior once.

Tag direction:

- `TagDefinition` is code: behavior, schema, defaults, allowed scopes, hooks.
- `TagInstance` is data: tag id plus parameters.
- `SourceTag` is an authoring macro that compiles to tag instances.

## Progression And Quest Direction

- Quest/progression is a stage machine, not a DAG.
- An active quest has exactly one `current_stage`.
- Stage definitions have no lifecycle.
- Stage `finish_conditions` are all required.
- No native optional objectives.
- No native `min/max`, `at_least`, choice groups, or parallel tracks.
- Branching is only one-of `next_stage`.
- Variable choices do not create branches.
- Complex optional, hidden, or parallel content should be independent quests connected by conditions.
- Final stage completion enters `reward_pending`.
- Rewards must be resolved before `completed` or `waiting_refresh`.
- Repeatable quests use `refresh_at` after reward resolution.
- There is no `auto_on_condition`; use `auto_on_event` with guard conditions.

## Git

- This is a single-person project; default to the current branch.
- New branches are allowed when they clearly help isolate risky or large work. Do not over-branch.
- Commit only when the user asks.
- After a requested commit, pushing is allowed when upstream/remotes are configured or GitHub sync is desired.
- Follow the repository style: Chinese plain commit messages.
- Before committing, inspect status, diff, and recent log.
- Never amend, force-push, or rewrite history unless explicitly asked.

## Agent Guardrails

- If unsure, ask the user instead of inferring from stale docs.
- Do not implement unapproved large systems.
- Do not revive discarded designs from old docs or conversation history.
- Do not expand non-TUI UX by default.
- Do not put executable logic in data.
- Do not parse item display names on the server.
- Do not weaken tests.
