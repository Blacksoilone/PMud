# Exit Tag Implementation Plan

**Goal:** Replace temporary room exit maps with room-local items carrying an `exit` tag, infer the eight planar standard directions from canonical inner names, and project those exits to movement, room observation, and the existing minimap.

**Architecture:** Content authors place ordinary items in rooms and attach `exit(target_room_id)`; compilation resolves the canonical inner name into an optional typed direction. The world resolves movement through exit items, while presentation sends visible exit direction/target projections to the client. The minimap consumes only the eight planar directions in this first slice.

**Tech Stack:** Go, existing content compiler/world/session/protocol/TUI packages, test-first development.

## Tasks

- [x] Add source and compiled exit-tag types; reject missing targets, unknown targets, and duplicate standard directions in one room.
- [x] Migrate world construction, movement, and room observation from `Room.Exits` to exit items.
- [x] Extend room protocol projection with direction-to-target room data while preserving the existing direction list.
- [x] Feed the eight planar neighbors into the current minimap; ignore up/down and future height-changing names.
- [x] Migrate tutorial content, remove temporary room exit maps, and run full diagnostics, lint, and race tests.
