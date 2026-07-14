# Tag Definition and Instance Design

## Purpose

The tag system separates content data from behavior code. Content data never contains executable logic, even for one-off effects. A content object can only declare tag instances and provide parameters. All behavior lives in tag definition files written as Go code, or later another code format such as Lua.

This design updates the earlier runtime-tag model from a plain string set into a typed call model:

```text
SourceTag -> compiled TagInstance -> TagDefinition handler
```

## Core Terms

`TagDefinition` is code. It declares a stable tag id, parameter schema, defaults, allowed scopes, supported rule contexts, and behavior handlers. A definition is written once and registered in a tag registry.

`TagInstance` is data. It appears on an entity or part and contains only the tag id plus parameter values. It is equivalent to calling a predefined behavior with arguments, but the call body is never stored in content data.

`SourceTag` is an author-facing macro. It can expand into one or more tag instances during content compilation. Source tags do not survive as runtime behavior unless they intentionally expand into a runtime tag instance with the same id.

## Data and Code Boundary

Content data may declare this:

```json
{
  "tag_instances": {
    "effect.undead_slayer": {
      "multiplier": 1.5
    }
  }
}
```

Content data must not declare this:

```json
{
  "tag_instances": {
    "effect.once": {
      "on_attack": "target.hp -= 999"
    }
  }
}
```

Even if an effect is used by one item only, it still gets a tag definition in code. Data calls behavior; data does not define behavior.

## Recommended Approach

Use schema-owned tag definitions.

Each tag definition owns its parameter contract because the tag definition is the behavior. The compiler validates every tag instance against the registered definition before producing runtime content. This catches missing fields, unknown fields, type errors, scope errors, and invalid values before play.

Alternative approaches were rejected:

- A content-side tag catalog would split behavior from its parameter contract and allow code/data drift.
- No schema would be faster at first, but typos and bad values would surface late during play.

## Runtime Flow

Runtime dispatch is generic. There should not be a central `if tag == ...` behavior switch.

The dispatcher performs the same generic sequence for every action context:

1. Determine which entity or part scopes are relevant to the action.
2. Collect tag instances from those scopes.
3. Look up each instance id in the tag registry.
4. Call the definition hook for the current context with the instance data.

The dispatcher knows how to route. It does not know what `effect.undead_slayer`, `property.soft`, or `visual.silver` mean.

## Part Scope

Parts are meaningful tag scopes, not just internal storage. A rule must explicitly decide which part scope it reads. Tags do not automatically bubble from a part to the item root.

For example, an attack with a sword should read the contact part, such as the blade. If only the hilt has a silver or undead-slayer tag instance, that should not grant the blade an undead damage bonus.

First-version gameplay does not need free player-driven part assembly or replacement, but parts can still be visible and explainable to the player.

## Source Tag Expansion

Source tags remain useful as author-facing macros:

```json
{
  "source_tags": ["silver"]
}
```

The compiler can expand this into typed tag instances:

```json
{
  "tag_instances": {
    "visual.silver": {},
    "effect.undead_slayer": { "multiplier": 1.5 },
    "property.soft": { "hardness": 3 }
  }
}
```

Expansion is deterministic. Cycles are compile errors. Unknown tag definitions are compile errors. Duplicate tag instances with conflicting data are compile errors unless the tag definition explicitly declares a merge rule.

## First-Version Boundary

The first playable version does not need to implement this full tag engine. Its priority remains the data-driven tutorial loop, aliases, ambiguity handling, item descriptions, inventory, movement, and core commands.

This design only fixes the direction for the later tag system: runtime tags are typed instances that invoke registered behavior definitions; content data never embeds executable behavior.
