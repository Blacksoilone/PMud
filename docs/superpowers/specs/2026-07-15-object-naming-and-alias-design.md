# Object Naming And Alias Design

## Status

Approved for design. This spec defines object naming, display, input aliases, and short handle syntax for the current Chinese/English MUD. It intentionally does not define a general i18n system.

## Goals

- Show players a concise, readable object name that exposes the intended Chinese name and English root name.
- Let players infer common input aliases from what they see without listing every alias.
- Keep object parsing authoritative on the server because object identity depends on dynamic world context.
- Reserve a simple short handle syntax for disambiguating repeated visible objects.
- Avoid spaces and `.` in object input names so command parsing remains simple.

## Non-Goals

- No full i18n resource model.
- No locale-keyed `zh` / `en` object name fields.
- No client-side object resolution.
- No natural-language parser.
- No recursive attribute comparison for disambiguation.

## Object Name Fields

Each named object has three naming concepts.

### `display_name`

`display_name` is the primary player-facing Chinese name.

Rules:

- It is shown before the English root name.
- It is accepted as an object input phrase by exact text match.
- It must not contain spaces.
- It must not contain `.` because `.` is reserved for short handles.
- It is not used to infer pinyin automatically.

Example:

```text
display_name = 铁剑
```

### `inner_name`

`inner_name` is the stable English root name.

Rules:

- It is shown in parentheses after `display_name`.
- It may contain spaces for natural display.
- The space-containing form is not accepted as object input.
- Input aliases are derived by removing spaces and accepting `-` / `_` separators.
- Matching is case-insensitive for ASCII letters.

Example:

```text
inner_name = iron sword
accepted input = ironsword, iron-sword, iron_sword, IRON-SWORD
rejected input = iron sword
```

### `input_aliases`

`input_aliases` are author-provided extra input names.

Rules:

- They are not shown by default.
- They are for pinyin, common nicknames, abbreviations, or other non-derived input forms.
- They must not contain spaces.
- They must not contain `.`.
- Matching is case-insensitive for ASCII letters.
- `-` and `_` are accepted as equivalent separators for alphabetic aliases.

Example:

```text
input_aliases = [tiejian]
accepted input = tiejian, tie-jian, tie_jian, TIE_JIAN
```

## Display Format

The default object display format is:

```text
display_name（inner_name）
```

Example:

```text
铁剑（iron sword）
```

The display string is not the same as an input string. In particular, the space-containing `inner_name` is for display only.

## Input Name Normalization

Object input phrases are normalized on the server during object resolution.

Rules:

- Object input phrases must not contain spaces.
- `.` is reserved for short handles and is not allowed inside normal names or aliases.
- ASCII case is folded.
- For `inner_name` and `input_aliases`, `-` and `_` are equivalent to no separator.
- `display_name` is accepted by exact text match, not by pinyin or other inferred transforms.

For `铁剑（iron sword）` with `input_aliases = [tiejian]`, accepted object phrases include:

```text
铁剑
ironsword
iron-sword
iron_sword
tiejian
tie-jian
tie_jian
```

Rejected object phrases include:

```text
iron sword
铁.剑
tiejian.abc
```

## Short Handles

When the server finds multiple visible or actionable objects that match the same object phrase, it may return candidates with short numeric handles.

The main handle syntax is:

```text
name.number
```

Examples:

```text
tiejian.1
铁剑.2
ironsword.1
```

Rules:

- `.` is the only primary handle separator.
- The suffix after `.` must be a positive decimal number.
- Handles are temporary object references, not world IDs.
- Handles are scoped to the player/session and the current server-provided candidate context.
- The client displays handles but does not interpret the object behind them.
- The server resolves handle references.

## Client And Server Boundary

The client handles command aliases and local syntax conveniences only.

Client responsibilities:

- Command aliases such as `take -> get`, `x -> examine`, `i -> inventory`, and movement shorthands.
- Future player-custom command aliases.
- Rendering object names and server-provided candidate handles.

The client must not:

- Resolve object names to object IDs.
- Resolve `display_name`, `inner_name`, or `input_aliases`.
- Maintain dynamic world object lookup tables for command resolution.
- Decide object ambiguity.

Server responsibilities:

- Resolve object phrases in the context of the command, player, room, inventory, container, actor, or other action domain.
- Apply `display_name`, `inner_name`, and `input_aliases` matching rules.
- Detect ambiguity only within the relevant action domain.
- Generate and resolve short handles.
- Return ambiguity feedback when the user must choose a specific object.

## Ambiguity Principle

Object ambiguity is contextual.

For example, `get shizi` should only consider objects that can be directly obtained from the current command domain. An identically named object elsewhere in the world must not create ambiguity.

The server should not recursively inspect arbitrary object internals to produce smart difference summaries. If multiple candidates need player selection, the server should provide short handles and normal display names. Players can examine candidates by handle if they need more information.

## Current Implementation Gap

The current codebase still uses `NameKey` and `Aliases` in content data, and the client currently resolves item names and aliases locally. That behavior is temporary and conflicts with this design.

Future implementation should migrate toward:

- Content fields equivalent to `display_name`, `inner_name`, and `input_aliases`.
- Server-side object phrase resolution.
- Client-side command aliasing only.
- Rendering that displays `display_name（inner_name）`.

## Example

Source concept:

```text
id = item.iron_sword
display_name = 铁剑
inner_name = iron sword
input_aliases = [tiejian]
```

Display:

```text
铁剑（iron sword）
```

Accepted input phrases:

```text
铁剑
ironsword
iron-sword
iron_sword
tiejian
tie-jian
tie_jian
铁剑.2
```

Rejected input phrases:

```text
iron sword
iron.sword
tiejian.abc
```
