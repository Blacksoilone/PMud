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
- `docs/superpowers/specs/2026-07-15-object-naming-and-alias-design.md`

## Project Shape

- This is a Go MUD project.
- The server is authoritative for world state, rules, and actions.
- The TUI is the primary product surface.
- Ordinary stdin/terminal mode is debug/compatibility only. Do not expand its UX unless the user asks.

## Development Workflow

- Use TDD for behavior changes: write the failing behavior test first, then implement.
- Test behavior boundaries, not every file or private helper.
- Prefer useful vertical slices, but optimize for delivering the first playable version of the big feature.
- Small slices are a means to an end: they exist to build foundations or clean up after a larger feature lands.
- Use small cuts when they are necessary, not when they are merely smaller.
- If the foundations are ready, do not keep postponing the larger feature.
- Refactoring has priority over feature work when existing design debt would make the larger feature harder or messier.
- Do not build major features on top of known historical baggage when that baggage is directly in the feature path.
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
- Test files are exempt from normal source LOC split triggers. Prefer keeping tests for the same behavior area together; split tests only when they mix unrelated behavior areas or become hard to navigate.

## Package Boundaries

- `internal/command`: command parsing and canonicalization.
- `internal/client`: TUI/client state, command aliases, rendering support, local static commands.
- `internal/session`: server command dispatch from canonical commands to world actions.
- `internal/world`: authoritative world graph, state, and rules. No text command parsing.
- `internal/content`: content source, compiler, server snapshot, client catalog projection.
- `internal/presentation`: server event to wire protocol.
- `internal/client/render`: wire protocol to player-facing text.

## Exit Direction

- Exits should eventually be first-class world entities/objects so they can be created, destroyed, hidden, locked, or transformed.
- Temporary room exit maps are acceptable during early implementation.
- Do not extend room exits into a separate permanent exit-specific rule system.
- Do not treat exits as normal inventory items in rendering.

## Command System

- TUI input goes through `internal/command.ParseClientInput`.
- The client handles fixed command aliases and future player-custom command aliases.
- The client must not resolve object display names, object aliases, or object ids.
- Object phrases are sent unresolved as part of canonical wire commands.
- The server parses canonical wire commands with `internal/command.ParseServerInput`.
- The server resolves object phrases in the current action context and reports object ambiguity.
- Current client-side item alias resolution is temporary and should migrate server-side.

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
- Agents may commit whenever current progress is worth saving.
- Agents may push committed progress to the current repository's current branch when an upstream/remotes are configured or GitHub sync is desired.
- Before staging, inspect status, diff, recent log, remotes, and `.gitignore`.
- Gitignore first: any file that should not be added must be covered by `.gitignore` before staging other files.
- Follow the repository style: Chinese plain commit messages.
- Never push to a different repository or remote than this current checkout's configured repository.
- Never amend, force-push, or rewrite history unless explicitly asked.

## Agent Guardrails

- If unsure, ask the user instead of inferring from stale docs.
- Do not implement unapproved large systems.
- Do not revive discarded designs from old docs or conversation history.
- Do not expand non-TUI UX by default.
- Do not put executable logic in data.
- Do not parse item display names on the client; object phrase resolution belongs server-side.
- Do not weaken tests.
