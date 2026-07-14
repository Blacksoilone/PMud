# Simplified Progression and Quest System Design

## Purpose

This document records the agreed long-term design for the MUD progression system. It is a design specification, not an implementation plan.

The system should support main quests, side quests, hidden quests, repeatable quests, discoveries, lightweight choices, rewards, and current-task UI without becoming a general dependency graph engine. The core rule is: keep the task system simple, and express complexity by composing independent quests, marks, world state, dialogue, items, and NPC checks.

## Core Principle

The system is a stage machine, not a DAG.

A quest has one current stage while active. A stage has conditions. All finish conditions must be satisfied before the stage ends. When a stage ends, the quest immediately moves to the next stage, a branch target, or `reward_pending` if the quest has no next stage.

No native support exists for optional objectives, min/max choice groups, parallel child tracks, arbitrary dependency graphs, or automatic UI summarization from graph structure.

## Terms

### Progression

The umbrella system for player progress entries. It manages visibility, activation, current stage, local progress, reward pending state, completion summaries, repeat refresh, and lightweight variables.

### Quest / ProgressEntry

A single progress entry. Main story, side quest, hidden quest, repeatable task, tutorial task, and discovery-like entries can all use this shape.

### Stage

A stage is one step in a quest stage machine. A quest in `active` lifecycle has exactly one `current_stage`. Stage definitions do not have their own lifecycle. A non-current stage is inert.

### Condition

A condition is a server-side rule used by quest activation, stage finish, stage failure, or quest failure. A stage can have multiple finish conditions, and all must be satisfied.

### Branch

A narrow one-of stage transition. A branch only decides which next stage follows the current stage. It does not support parallel branches, multiple selected paths, min/max completion, or optional sub-objectives.

### Variable Choice

A choice that does not change the next stage. It records a variable, mark, relationship value, or text parameter used by later content.

### Mark

A persistent player fact, memory, permission, unlock, or qualification. Examples include visiting a city, hearing a piece of lore, joining a faction, or earning a trusted status with an NPC.

### RewardEffect

A declarative reward effect triggered by quest/stage events. Data names the effect kind and parameters; Go code implements the effect behavior once.

### RefreshAt

The next time a repeatable quest becomes available again after its reward has been handled.

## Quest Lifecycle

Quest lifecycle is quest-level only:

```text
hidden
unlocked
active
reward_pending
completed
waiting_refresh
retry_wait
```

`hidden` means the player does not know the quest exists.

`unlocked` means the quest is known or available but not active.

`active` means the quest is running and has one `current_stage`.

`reward_pending` means the final stage has completed, but rewards have not been fully handled or claimed.

`completed` means a one-time quest is finished and all required reward handling is complete.

`waiting_refresh` means a repeatable quest has completed and handled rewards, and is waiting for its next `refresh_at`.

`retry_wait` means an active run failed or was abandoned and is waiting before it can return to `unlocked`.

## Stage Rules

A stage definition can contain:

```text
id
```

Stage rules:

1. A quest has exactly one current stage while active.
2. Stage definitions have no lifecycle of their own.
3. A stage's finish conditions are all required.
4. When the current stage finishes, the quest immediately moves to the next stage or selected branch target.
5. If there is no next stage, the quest moves to `reward_pending`.
6. If the design needs the player to confirm moving on, that confirmation is a finish condition of the current stage.
7. Stage-local progress is cleared when leaving the stage.

Example:

```text
stage.prepare_departure
  text: Prepare for the journey and speak with the courier.
  finish_conditions:
    - dialogue_choice(courier, confirm_departure)
  next_stage: stage.on_road
```

Optional preparation such as buying food, changing horses, hiring guards, or buying a map is not represented as stage requirements unless it is truly mandatory. The courier dialogue can inspect those world facts and set variables, marks, risks, or different text outcomes.

## Finish Conditions Are Always All

The task system does not support native `any`, `at_least`, `min/max`, or optional stage requirements.

If a stage lists these conditions:

```text
[ ] Kill 10 spiders
[ ] Collect 5 herbs
```

both must be satisfied.

If content wants player choice or optional preparation, model it outside the stage requirement list:

- use NPC/world interaction checks;
- use independent hidden or side quests;
- use marks, variables, items, and relationship state;
- use a narrow one-of branch only when the next stage truly changes.

## Branch Rules

Native branching is intentionally narrow.

A branch can only choose one next stage when the current stage finishes:

```text
stage.decide_prisoner
  branch release -> stage.release_path
  branch report  -> stage.report_path
  branch kill    -> stage.kill_path
```

Once one branch is chosen, the other branch targets are not taken.

Allowed:

```text
A -> B1 -> D
A -> B2 -> D
```

The player only travels through one branch. A later explicit merge is allowed by pointing multiple branch paths to the same later stage.

Not allowed:

```text
A activates B1 and B2 in parallel
D requires B1 and B2
```

Parallel or optional structure should be expressed with independent quests and conditions, not native branches.

## Variable Choices

Not every choice is a branch.

A choice is a branch only if it changes `next_stage`. If the choice only changes later text, NPC address form, tone, relationship, reward parameters, or small content differences, it is a variable choice.

Examples:

```text
Branch choice:
  release prisoner -> stage.help_prisoner_escape
  report prisoner  -> stage.report_to_guard

Variable choice:
  NPC calls player "young hero" / "traveler" / "sir"
  set quest variable address_form
  continue to the same next stage
```

Variable lifetimes can be:

```text
stage-local
quest-local
quest-summary
persistent
```

Only variables needed after a stage or quest should survive cleanup.

## Independent Quests Express Complexity

The system does not support native optional objectives or parallel quest tracks inside one quest. Use independent quests instead.

Example: main quest `A` reaches stage `d`, which opens hidden quest `B`. Main quest `A` later reaches stage `f`, whose finish conditions include `quest_completed(B)`.

```text
Quest A:
  stage a -> b -> c -> d -> e -> f

Quest B:
  activation: quest_stage_is(A, d)
  failure: quest_stage_is(A, f)
  failure_penalty: none

Quest A stage f:
  finish_conditions:
    - quest_completed(B)
```

This can produce the player experience of a dual-line narrative without making the core quest model support parallel child tracks.

Rules for quest-to-quest relationships:

1. A quest does not own another quest.
2. Quests can read other quests' summary state through conditions.
3. Quests should not mutate another quest's internal stage directly.
4. Hidden, optional, and side content should be independent quests with their own lifecycle.
5. Main quests can depend on hidden quests by checking completion state.

## Activation Policy

Quest activation has only these policies:

```text
manual_accept
auto_on_event
always_active
```

`manual_accept` starts when the player explicitly accepts or starts the quest.

`auto_on_event` starts when a declared event occurs and guard conditions pass. There is no `auto_on_condition` policy. Conditions are guards checked only when the event fires.

`always_active` starts immediately for initial tutorial or starting main content.

Repeat refresh is not its own activation policy. A repeatable quest in `waiting_refresh` registers a `refresh_at`; when due, the system emits a refresh event. The quest can return to `unlocked`, or use `auto_on_event` if it should start immediately.

The progression system must not poll every quest every tick.

## Conditions and Events

Conditions can inspect current state or stage-local progress, but checks happen in response to explicit events or player actions.

Common condition types:

```text
event_trigger
event_counter
has_mark
has_item
quest_lifecycle_is
quest_stage_is
quest_completed
variable_equals
world_state_check
```

For ordinary MMO behavior such as killing monsters or gathering items, use event counters that count only after the quest/stage is active.

Global player statistics may exist, but future quests should not automatically complete from historical statistics unless a condition explicitly reads those statistics.

## Local Progress and Cleanup

Active stages can store local progress:

```json
{
  "quest.herbalist": {
    "lifecycle": "active",
    "current_stage": "gather_supplies",
    "local_progress": {
      "spider_kills": 7,
      "herbs_collected": 3
    }
  }
}
```

When the stage ends, this local progress is deleted.

When the quest completes, stage detail is compressed into a quest-level summary:

```json
{
  "quest.herbalist": {
    "lifecycle": "completed",
    "completed_at": 12345,
    "reward_claimed_at": 12360
  }
}
```

The player save must not store quest definitions, stage definitions, reward definitions, or completed stage internals.

## Marks

Marks are persistent facts that have long-term meaning:

```text
visited.changan
heard_lore.changan_customs
met.npc.lao_chen
trusted_by.gate_guard
unlocked.area.east_gate
joined.faction.merchant_guild
```

Marks are not task progress. A stage may check marks as conditions, but marks should not be used to model short-lived counters such as killing 10 spiders for a current task.

## Rewards

Rewards are effects, not hard-coded task logic.

Reward effects can include:

```text
grant_item
grant_money
grant_exp
grant_mark
set_variable
unlock_area
grant_title
change_reputation
```

Reward behavior is implemented in code. Data only declares effect kind and parameters.

Completion and reward claiming are separate. Final stage completion moves a quest to `reward_pending`. The quest cannot move to `completed`, `waiting_refresh`, or a new run until rewards are resolved.

This handles inventory-full cases:

```text
active -> reward_pending
player makes inventory space
reward_pending -> completed
```

For repeatable quests:

```text
active -> reward_pending -> waiting_refresh -> unlocked
```

## Repeatable Quests

Repeatable quests use `refresh_at`.

After a repeatable quest's rewards are resolved:

1. local progress is already cleared;
2. `completion_count` increments;
3. `last_completed_at` is recorded;
4. the system computes `refresh_at`;
5. lifecycle becomes `waiting_refresh`.

Example save state:

```json
{
  "quest.daily.herbs": {
    "lifecycle": "waiting_refresh",
    "completion_count": 5,
    "last_completed_at": 12345,
    "refresh_at": 13000
  }
}
```

When `refresh_at` is due, the quest returns to `unlocked` unless content explicitly starts it through `auto_on_event`.

## Failure and Abandonment

Quests can fail or be abandoned, but MMO content should not default to permanent punishment.

Failure or abandonment:

1. applies only during `active`;
2. terminates the current run;
3. clears current stage local progress;
4. can trigger light outcome effects;
5. moves to `retry_wait` or `unlocked`.

Examples of light penalties:

```text
short retry wait
loss of deposit
teleport to checkpoint
cleanup of temporary task item
increment fail_count or abandon_count
```

Permanent consequences must be explicit and should not be default for mainline MMO content.

`reward_pending` cannot be abandoned to roll back completion.

## UI Rules

The UI displays the quest name, current stage text, and current stage conditions.

It does not display optional internal objectives because the system has no native optional objectives.

If content wants optional preparation, the UI should show a broad stage goal such as:

```text
Prepare for departure and speak with the courier.
```

NPC dialogue, items, marks, and world interactions can suggest or inspect preparation choices.

The task UI should not expose hidden independent quests until those quests become visible through their own lifecycle.

## Data and Code Boundary

Quest data declares ids, text keys, stages, conditions, effects, branches, variables, lifecycle policy, and refresh rules.

Quest data must not contain executable scripts. Condition kinds and reward effect kinds are implemented in code once and called by data with parameters.

This follows the same principle as the tag system: data invokes predefined behavior; behavior does not live inside content data.

## Hard Rules

1. Quest is a stage machine, not a DAG.
2. Quest active state has exactly one current stage.
3. Stage definitions have no lifecycle.
4. Stage finish conditions are always all required.
5. Optional objectives and parallel tracks are not native features.
6. Complex optional or parallel content is expressed as independent quests.
7. Branching is only one-of next stage.
8. Variable choices do not create branches.
9. Final stage completion always reaches `reward_pending` before final completion or repeat refresh.
10. Rewards must be resolved before completion or repeat refresh.
11. Repeatable quests use `refresh_at` after reward resolution.
12. There is no `auto_on_condition`; automatic activation is event-triggered with guard conditions.
13. Player saves store sparse runtime state and summaries, not quest definitions or completed stage internals.
14. Data contains no executable code.
