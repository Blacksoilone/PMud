# Right-HUD TUI Visual Design

This spec defines the first polished TUI layout for the MUD. It is a visual and interaction contract for implementation. It intentionally does not design RPG stats, equipment rules, map generation, or alternate layout presets.

## Goals

- Ship a serious terminal page with clear permanent regions, elegant borders, rich but disciplined styling, popup overlays, and scrollable log history.
- Keep the first layout fixed to a right-side vertical HUD.
- Keep permanent data compact; detailed data belongs in popups.
- Preserve the bottom input line at all times. Popups never cover it.
- Support mouse wheel scrolling: no popup means scroll log; popup open means scroll popup.

## Terminal Size Contract

The user's display uses strict full-width/half-width 2:1 metrics.

- Minimum supported size: `54 x 26` full-width units, equivalent to about `108 x 26` terminal cells.
- Comfortable default: `64 x 32` full-width units, equivalent to about `128 x 32` terminal cells.
- Ideal range: `72-78 x 36-42` full-width units, equivalent to about `144-156 x 36-42` terminal cells.
- Below `54 x 26`, the TUI should show a clear terminal-too-small message and avoid broken layout.

First version resizing rules:

- Width changes resize only the left main column and the bottom input line. The right HUD keeps a fixed width.
- Height changes resize the left log pane and the right tracked-quest pane.
- Room/visible-objects, minimap, and status placeholder keep fixed heights.
- Later versions may switch presets at thresholds, but first version does not.

## Layout Skeleton

Default layout: right vertical HUD.

```text
┌─ 房间 / 可见物 ───────────────────────────────┬─ 小地图 ─────────────┐
│ 练习场入口                                    │        北             │
│ 这里是一处安静的练习场入口……                  │        │              │
│ 可见物: 旧油灯（old lantern）                 │ 西 ─ 当前 ─ 东        │
│ 出口: 北                                      │        │              │
├─ 日志 ────────────────────────────────────────┤        南             │
│ 你拿起了旧油灯（old lantern）。               ├─ 状态 ───────────────┤
│ 任务更新: 走进院子                            │ 状态系统未开放        │
│ 你向北走去。                                  │                      │
│                                               ├─ 当前任务 ───────────┤
│                                               │ 初入练习场            │
│                                               │ 阶段: 走进院子        │
│                                               │ 目标: 查看练习木剑    │
├───────────────────────────────────────────────┴──────────────────────┤
│ > examine practice-sword                                              │
└───────────────────────────────────────────────────────────────────────┘
```

Layout constants for first implementation:

- Input pane: fixed bottom row with border, one line of visible input.
- Right HUD width: fixed. Recommended first value: `18` full-width units (`36` terminal cells) including borders.
- Room pane height: fixed. Recommended first value: `7` rows including border.
- Minimap pane height: fixed. Recommended first value: `7` rows including border.
- Status placeholder height: fixed. Recommended first value: `5` rows including border.
- Log pane height: fills left-column remainder above input.
- Tracked quest pane height: fills right-column remainder above input.

## Visual Language

Use a calm dark terminal surface with restrained color. The TUI should look like a game console, not a generic dashboard.

Palette intent:

- Base background: terminal default or near-black.
- Pane border: muted slate/gray.
- Focus or active border: warm amber or soft cyan.
- Main titles: bold, high-contrast.
- Secondary labels: dim gray.
- Player/action text: normal foreground.
- Important system or quest updates: amber/yellow accent.
- Errors: red accent, bold only for the important word.
- Item names: cyan or green accent, not over-saturated.
- Directions/exits: blue/cyan accent.

Use box characters deliberately:

- Permanent panes use single-line borders: `┌ ┐ └ ┘ ─ │ ├ ┤ ┬ ┴ ┼`.
- Active/focused pane title may use a brighter color but keeps single-line border.
- Popups use double-line borders: `╔ ╗ ╚ ╝ ═ ║` to distinguish overlay mode.
- Avoid noisy decorative glyphs in body text. Decoration belongs in borders, titles, scroll indicators, and status lines.

## Permanent Pane Appearance

### Room / Visible Objects Pane

Purpose: current room identity, concise room description, visible objects, and exits.

Appearance:

- Title: `房间 / 可见物`, bold.
- Border: muted single-line border.
- Room name: first content line, bold.
- Description: wrapped CJK-aware text, normal color.
- Visible objects line: label `可见物:` dim; item names accent colored.
- Exits line: label `出口:` dim; directions accent colored.
- No scrolling in first version. If content overflows, truncate with `…` and rely on log/detail commands for longer text.

### Log Pane

Purpose: the largest history pane. It shows everything not permanently represented elsewhere: action feedback, system messages, quest progress notifications, errors, NPC speech, combat text, and debug-compatible messages.

Appearance:

- Title: `日志`, bold.
- Border: muted single-line border. When actively scrolled away from bottom, title includes a subtle indicator such as `日志 ↑历史`.
- Content: chronological lines, CJK-aware wrapped.
- Quest progress lines: amber accent on the label or marker.
- Error lines: red accent.
- NPC speech: optional quoted style, e.g. dim speaker name and normal speech.
- Scroll indicator: right side of title or bottom-right inside pane, such as `37%` or `底部`.
- Keyboard and mouse wheel both scroll log when no popup is open.
- Log scroll never affects other panes.

### Minimap Pane

Purpose: compact local map only. Large map belongs in a popup.

Appearance:

- Title: `小地图`, bold.
- Border: muted single-line border.
- Content: centered small ASCII/CJK-safe direction map.
- Current room marker: highlighted `当前位置` or `◎` if width is tight.
- Exits: accent colored direction labels or lines.
- Fixed height; no scrolling.

First version may use a simple direction sketch from known exits rather than a true graph map.

### Status Placeholder Pane

Purpose: reserve the personal status area without designing RPG systems early.

Appearance:

- Title: `状态`, bold but dimmer than active gameplay panes.
- Border: muted single-line border.
- Body: one concise placeholder line, e.g. `状态系统未开放`.
- Fixed height; no RPG stats, equipment, held item, hunger, thirst, combat state, or attributes in first version.

### Tracked Quest Pane

Purpose: compact summary of the currently tracked quest only.

Appearance:

- Title: `当前任务`, bold.
- Border: muted single-line border.
- Shows only:
  - Quest name
  - Stage name
  - Stage objective
- Does not show detailed description, rewards, all conditions, or task-switching controls.
- Does not allow direct switching tracked quest from the permanent pane.
- Height is flexible and absorbs right-HUD vertical resizing.
- If content overflows in first version, truncate with `…`. Detailed task data belongs in a popup.

### Input Pane

Purpose: command entry line, always visible.

Appearance:

- Bottom pane with single-line border.
- Content begins with a visible prompt: `> `.
- Input text uses normal foreground; cursor styling should remain compatible with CJK IME.
- Never covered by popups.
- Force-redraw shortcut is available here even when no popup is open.

## Popup System

Popup behavior:

- Exactly one popup may be active at a time.
- Opening another popup while one is active switches to the new popup, replacing the old one.
- Popup overlays the content area above the input pane only. It never covers the input pane.
- Background panes are dimmed while a popup is active.
- Popup owns keyboard scrolling and mouse wheel scrolling.
- Bottom input line remains visible. First version may freeze normal command submission while a popup is active, except for popup commands and popup-switch commands.
- `Esc` closes the current popup.

Popup sizing:

- Popup size is computed from terminal/content area size.
- Width: `70%` of terminal cell width, clamped between `64` and `110` terminal cells.
- Height: `70%` of content area height above input, clamped between `10` and `28` rows.
- Popup is centered horizontally and vertically within the content area above input.
- If the terminal is near minimum size, use the maximum area available above input while preserving at least a one-cell margin where possible.
- Popup content area scrolls internally when content exceeds popup height.

Popup border and style:

- Double-line border.
- Title centered or left-aligned with accent color.
- Footer line shows local operations, e.g. `[Esc] 关闭  [↑↓/滚轮] 滚动`.
- Body uses normal foreground with labels dimmed and important values accented.

### Help Popup

Purpose: local UI and command help. This popup may use local client text because it describes UI controls, not authoritative world state.

Appearance:

- Title: `帮助`.
- Sections: `基础命令`, `界面操作`, `弹窗操作`.
- Commands are highlighted; descriptions normal.
- Scrollable.

Behavior:

- Open with `?` or `F1`.
- `Esc` closes.
- Mouse wheel and arrow keys scroll popup.

### Inventory Popup

Purpose: non-permanent inventory view. Content must come from server-authoritative inventory data.

Appearance:

- Title: `背包`.
- Item rows show display name, inner name, and concise state if available.
- Empty inventory shows a dim empty-state line: `背包为空`.
- Scrollable.

Behavior:

- Open with `i` / `inventory` through a server inventory request, then display from client inventory region/state.
- `Esc` closes.
- Mouse wheel and arrow keys scroll popup.
- If help is open and inventory is requested, switch to inventory popup.

Future popups, not first-version requirements:

- Quest detail popup
- Character detail popup
- Equipment popup
- Large map popup

## Mouse And Keyboard Rules

Mouse first version:

- Enable Bubble Tea mouse cell motion.
- No popup: any mouse wheel position scrolls the log.
- Popup active: any mouse wheel position scrolls the active popup.
- Do not hit-test panes for mouse-wheel targets in first version.
- Do not implement hover effects in first version.

Keyboard first version:

- `PageUp` / `PageDown`: scroll log when no popup; scroll popup when popup active.
- `↑` / `↓`: scroll popup when popup active. When no popup, keep existing command input behavior unless a focused-log mode is explicitly introduced.
- `?` or `F1`: open/switch help popup.
- `i` / `inventory`: request inventory data and open/switch inventory popup.
- `Esc`: close popup if open.
- Force redraw: `Ctrl+R`.

Force redraw behavior:

- `Ctrl+R` should force a full TUI redraw without changing game state, input buffer, log scroll position, or popup state.
- If a popup is active, redraw keeps the popup active.
- This is a local client operation; it should not send a command to the server.

## Data Boundaries

- Client owns layout, borders, colors, scroll offsets, popup identity, and redraw behavior.
- Server owns world-state data: inventory, quest detail, character detail, equipment, map data, room state, visible objects.
- Help popup may be client-local because it documents client UI controls.
- Inventory popup must be backed by the server inventory event/state.

## First-Version Non-Goals

- No alternate layout preset implementation.
- No left-HUD, top-HUD, compact preset, or threshold-based full layout switching.
- No RPG stat design.
- No equipment or character detail popup.
- No large map popup.
- No pane-specific mouse hit testing.
- No hover interactions.
- No popup stack.
- No direct tracked-quest switching from the permanent quest pane.
