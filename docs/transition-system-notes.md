# Transition 系统设计笔记

Transition 是受规则控制的位置变更。它覆盖普通出口，也覆盖死亡、主动脱困、回城符、渡船、车费旅行、陷阱、梦境、被捕、副本 rollback 等非普通走路的移动。

核心原则：

> Exit 是最常见、最静态、最便宜的 Transition。Transition 是更通用、更动态的位置变更机制。

## 为什么需要 Transition

传统出口系统通常表达固定房间连接：

```text
north -> town.road
enter gate -> city.inside_gate
climb ladder -> tower.second_floor
```

这适合大多数普通移动，但不足以表达：

```text
支付车费前往某个城市
乘坐渡船，目的地随航线和时间变化
通过同一个传送门，根据钥匙、状态或时间到达不同地点
死亡后进入某个死后世界
主动脱困回到安全区域
被强盗击败后关进某个牢房
触发陷阱后掉入坑底
睡在祭坛上进入梦境地图
副本失败后回到 safe_exit_place
房间被热加载删除后迁移到 fallback_place
```

如果这些都硬塞进固定出口，会出现大量特例出口，例如：

```text
enter_portal_if_has_red_key
enter_portal_if_has_blue_key
enter_portal_at_night
enter_portal_for_faith_x
enter_portal_for_quest_y
```

这会污染房间结构，也让规则难以复用。

## 统一模型

所有受控位置变更都可以进入同一条管线：

```text
Action / Rule / Event
  -> TransitionRequest
  -> TransitionResolver
  -> TransitionPlan
  -> PositionTransaction
```

### TransitionRequest

TransitionRequest 描述“为什么要移动”和“谁要移动”。

```text
TransitionRequest:
  actor_id
  trigger_kind
  source_place
  context_refs
  requested_destination?
  reason
```

trigger_kind 可以是：

```text
walk_exit
enter_portal
pay_travel_fare
board_ferry
death
escape
use_item
cast_spell
trap_triggered
captured
dungeon_rollback
invalid_location_recovery
```

### PlaceResolver

PlaceResolver 负责回答“去哪里”。

```text
PlaceResolver:
  Resolve(context) -> TransitionDestination
```

TransitionDestination 不一定总是单个房间。它可以是：

```text
fixed destination:
  resolver 已经确定唯一 PlaceID

candidate destinations:
  resolver 返回多个可选目的地，等待玩家选择或确认

blocked reason:
  当前条件不足，不能移动，并给出原因
```

常见 resolver：

```text
FixedPlaceResolver:
  固定目标房间

AreaFallbackResolver:
  当前区域安全点

PlayerBoundHomeResolver:
  玩家绑定点

AfterlifeResolver:
  根据地区、文明、信仰、种族和特殊状态选择死后世界

PortalTargetResolver:
  根据钥匙、时间、状态或任务决定传送门目标

FerryRouteResolver:
  根据渡船当前航线和班次决定目的地

CapturePlaceResolver:
  根据击败玩家的势力决定牢房

DungeonSafeExitResolver:
  副本 safe_exit_place
```

目的地选择有三种默认模式：

```text
系统自动选择:
  死亡进入死后世界
  escape 回到安全点
  陷阱掉入坑底
  副本 rollback 到 safe_exit_place

玩家显式选择:
  乘车去哪个城市
  传送阵选择哪个已解锁地点
  渡船选择可达港口

先给候选再确认:
  付费旅行
  高成本传送
  不可撤销或有明显风险的转移
```

原则：

```text
环境、死亡、陷阱、失败恢复:
  自动选择目的地

交通、付费旅行、主动传送:
  玩家显式选择或确认

高成本、不可撤销、有风险的转移:
  必须确认
```

这样可以避免把所有动态目的地都设计成命令特例。resolver 可以先产出候选列表，再由命令系统进入一次歧义/确认流程。

### TransitionPolicy

TransitionPolicy 负责回答“这次移动需要什么、会清理什么、到达后怎样”。

```text
TransitionPolicy:
  conditions
  costs
  cleanup
  interruption_rules
  arrival_state
  presentation
```

例子：

```text
DefaultWalkPolicy:
  检查出口可见、门未锁、actor 可移动

PaidTravelPolicy:
  检查路线开放
  扣车费或船票
  可选等待发车

RecallPolicy:
  检查当前房间是否禁用 recall
  消耗回城符
  清理部分临时状态

DeathTransitionPolicy:
  清理 combat 和临时收益
  进入 dead / spirit 状态
  移动到死后世界

DungeonRollbackPolicy:
  rollback run
  release reserved ticket
  discard pending rewards
```

TransitionPolicy 第一版应保持受限，不要变成脚本系统。它可以负责：

```text
conditions:
  额外进入条件，例如不能在战斗中回城、必须有船票、必须是 spirit 状态

costs:
  扣钱、消耗道具、消耗次数、消耗门票 reservation

cleanup:
  清理临时收益、打断战斗、释放跟随、清理副本内物品

arrival_state:
  到达后给予限制、状态或标记，例如 escape 后不能移动，死亡后进入 spirit 状态

presentation:
  输出移动前后文本
```

TransitionPolicy 不应负责：

```text
任意创建复杂奖励
修改无关系统
扫描全世界
执行循环逻辑
决定复杂剧情分支
```

也就是说：

```text
PlaceResolver:
  决定去哪儿

TransitionPolicy:
  决定能不能去、付什么代价、带走什么、到达后有什么状态

PositionTransaction:
  真正改变位置
```

### Cost Timing

Transition 的成本结算要区分三种模式。

```text
pre_pay:
  先扣费，再移动。
  不推荐默认使用，因为移动失败时需要退款。

commit_on_success:
  先验证能支付，移动成功后在同一事务里扣费。
  推荐作为即时 Transition 的默认模式。

reserve_then_settle:
  先 reserve，稍后结算。
  适合副本门票、长期旅程、等待发车的渡船、押镖、制作等有生命周期的玩法。
```

默认规则：

```text
即时 Transition:
  commit_on_success

有等待或生命周期的 Transition:
  reserve_then_settle

pre_pay:
  只有内容明确需要时才使用
```

这个规则与副本门票模型一致：进入时 reserve，完成时 consume，rollback 时 release。

### Failure Feedback

Transition 失败不应只是返回 false，而应返回结构化 blocker。移动失败本身也是游戏体验的一部分，尤其是门、渡船、传送阵、死后世界转生门等场景。

这些 blocker 不直接拼自然语言。它们应交给 Presentation 系统渲染；具体边界见 `presentation-system-notes.md`。

```text
TransitionResult:
  success
  destination?
  blockers
  candidates?

TransitionBlocker:
  code
  severity
  source
  public_message_key?
  debug_message?
  hint?
  reveal_policy
```

blocker 类型：

```text
hard_blocker:
  完全不能移动
  例如缺钥匙、缺船票、正在战斗、房间禁止 recall

soft_warning:
  可以移动，但有代价或风险
  例如这会消耗一张船票，进入后无法返回

candidate_lock:
  某个候选目的地不可用，但其它候选仍可选
  例如传送阵里某个城市未解锁
```

例子：

```text
missing_item:
  source: inventory
  required: ferry_ticket
  public: 你没有船票。
  hint: 你可以在码头售票人那里购买船票。

forbidden_state:
  source: actor
  state: in_combat
  public: 你正在战斗中，不能使用回城符。

attribute_out_of_range:
  source: actor
  field: age
  required: age >= 16
  public: 守卫只允许成年人进入。

forbidden_item:
  source: inventory
  item_tag: iron_weapon
  public: 镜门排斥铁器。
```

blocker 需要区分公开反馈和调试反馈。否则谜题、隐藏门和条件式传送很容易被系统剧透。

```text
public_message_key:
  给玩家看的说明

debug_message:
  给管理员、builder 或内容检查工具看的原因

reveal_policy:
  always
  if_known
  admin_only
  never
```

多 blocker 同时存在时，默认只显示一个最重要的公开原因，避免刷屏。

```text
普通移动命令:
  显示最高优先级 public blocker

检查命令，例如 inspect gate / routes / why:
  可以显示更多公开原因、候选目的地和提示

管理员/调试模式:
  可查看 debug blockers
```

原则：

> Transition 失败返回结构化 blockers。日常命令默认只显示最重要的公开原因；玩家主动检查时再展开更多信息；隐藏条件通过 reveal_policy 避免剧透。

### PositionTransaction

最终移动必须通过统一 PositionTransaction，保证位置唯一性和原子性。

```text
PositionTransaction:
  validate actor current position
  remove actor from source place
  add actor to destination place
  update location relation
  emit movement / arrival events
```

任何系统都不应该绕过 PositionTransaction 直接改位置。

## Exit 与 Transition 的关系

Exit 是房间局部的 TransitionRule。

```text
Exit = room-local TransitionRule
StaticExit = FixedPlaceResolver + DefaultWalkPolicy
DynamicExit = custom PlaceResolver + custom TransitionPolicy
```

底层执行统一为 Transition，但内容表达不必统一得很复杂。

普通出口保留简单写法：

```text
Exit:
  direction: north
  to: village.road
```

等价于：

```text
TransitionRule:
  trigger: command("north")
  resolver: FixedPlace(village.road)
  policy: DefaultWalkPolicy
```

这样内容作者写普通地图时不需要理解完整 Transition 系统。

## StaticExit

StaticExit 是最常见出口。

```text
Room: town.square
Exit north:
  to: town.north_road
```

可选属性：

```text
door
locked_by
hidden
one_way
movement_cost
required_state
blocked_message
```

特点：

```text
固定目标
规则简单
适合大多数房间连接
性能和可读性最好
```

## DynamicExit

DynamicExit 看起来仍然是一个出口，但目的地由 resolver 决定。

例子：同一个传送门，不同目标。

```text
Exit enter mirror:
  resolver: mirror_destination

mirror_destination:
  if actor has dream_mark -> dream_palace.entrance
  else if world time is night -> moon_corridor
  else -> dusty_room
```

例子：城门根据通行证进入不同地点。

```text
Exit enter gate:
  resolver: city_gate_destination_by_pass
  policy: check_city_pass
```

特点：

```text
同一个入口
目的地可变
可按时间、钥匙、状态、身份、任务、世界事件变化
适合传送门、梦境入口、特殊门、魔法镜、随机迷宫
```

## 非出口 Transition

有些位置变更不是房间方向或入口，不应伪装成出口。

### 支付车费

```text
Trigger:
  command("ride cart to <city>")

Resolver:
  selected_destination_from_cart_routes

Policy:
  require city in allowed_routes
  pay fare
  move actor to destination station
```

### 渡船

```text
Trigger:
  board ferry

Resolver:
  ferry_route_current_or_next_stop

Policy:
  require ticket or fare
  wait until departure if needed
  move passenger when ferry arrives
```

渡船也可以被建模为移动中的 Place：玩家先进入船，船到站后 Transition 把船上乘客移动到码头。

## Group Transition

有些 Transition 必须处理多个对象，例如渡船乘客、马车乘客、跟随者、押送目标、召唤物、副本 rollback 中的玩家和 run-local 对象。

这类场景不应被实现成“对每个对象独立执行一次完整 Transition”。否则会出现：

```text
费用重复扣除
resolver 每次算出不同目的地
中途失败导致一半对象移动、一半对象留下
presentation 重复刷屏
跟随、押送、乘客关系被打断
```

推荐模型是：一次 Transition 先生成一个 group plan，然后底层仍逐个对象执行 PositionTransaction。

```text
GroupTransitionPlan:
  trigger_actor
  participants
  destination / per-participant destination
  shared_policy
  execution_mode

execute:
  resolve participants
  validate required participants
  apply shared costs / cleanup
  for each participant:
    PositionTransaction(participant, destination)
  emit group transition result
```

### Execution Mode

不同 group transition 的失败策略不同，不应强行统一。

```text
all_or_nothing:
  所有必要对象都能移动，才执行。
  适合组队进入副本、双人机关门、需要全队确认的传送。

best_effort:
  能移动的移动，不能移动的留下并给出原因。
  适合普通跟随、队长走路、可选宠物跟随。

force_required:
  规则要求相关对象必须被处理，不能各自选择。
  适合渡船到港、马车到站、被押送 NPC、副本 rollback、死亡后的召唤物清理。
```

### Participant Role

participants 不应只是一组 actor id。它们应携带角色和失败策略。

```text
Participant:
  entity_id
  role:
    primary
    follower
    passenger
    escorted
    summoned
    prisoner
    vehicle
    contents
  required: true/false
  failure_policy:
    block_all
    leave_behind
    force_move
    cleanup
```

例子：普通跟随。

```text
primary: 玩家A, required=true
follower: 玩家B, required=false, failure_policy=leave_behind
pet: 狗, required=false, failure_policy=leave_behind
execution_mode: best_effort
```

例子：渡船到港。

```text
vehicle: ferry, required=true
passenger: 玩家A, required=true, failure_policy=force_move
passenger: 玩家B, required=true, failure_policy=force_move
contents: 船上物品, required=true, failure_policy=force_move
execution_mode: force_required
```

例子：副本 rollback。

```text
primary: dungeon_run, required=true
participant: 玩家, failure_policy=force_move_to_safe_exit
summoned: 召唤物, failure_policy=cleanup
run_local_item: 临时物品, failure_policy=cleanup
execution_mode: force_required
```

原则：

> Transition 是一次受控位置变更计划，可以有一个 primary actor，也可以有多个 participants。resolver 和 shared policy 执行一次；实际位置修改仍逐个通过 PositionTransaction 完成。

### 死亡

```text
Trigger:
  actor died

Resolver:
  AfterlifeResolver

Policy:
  clear combat and temporary rewards
  enter spirit state
  move to afterlife place
```

### 主动脱困

```text
Trigger:
  command("unstuck")

Resolver:
  FallbackPlaceResolver

Policy:
  clear exploitable temporary state
  move immediately
  apply post-escape restriction and cooldown
```

### 被捕

```text
Trigger:
  defeated_by_bandits

Resolver:
  CapturePlaceResolver

Policy:
  remove combat state
  confiscate selected items if rule says so
  move to prison cell
```

## 谁声明 Transition 规则

Transition 不是房间或人物的专属能力。房间、区域、人物、物品、能力和玩法实例都可以声明规则或修饰符。

### 房间 / 区域

适合声明环境规则：

```text
exits
dynamic exits
fallback_place
safe_respawn_place
afterlife_ref
recall_forbidden
escape_policy
trap_transition
on_enter_transition
```

### 人物

适合声明身份和绑定规则：

```text
home_place
faction_respawn_place
faith_afterlife_override
race_afterlife_override
death_protection_items
movement_lock
cannot_transition_flags
```

### 物品 / 能力

适合声明主动转移规则：

```text
recall_scroll -> bound home
teleport_spell -> selected target place
dream_key -> dream map
death_token -> lower-cost afterlife gate
```

### 玩法实例

适合声明局部规则：

```text
dungeon_safe_exit
death_inside_policy
escape_policy
logout_policy
rollback_transition
arena_exit
```

## 规则合成

TransitionResolver 应按上下文收集规则，而不是只查一个对象。

```text
collect actor rules
collect source place rules
collect area rules
collect item / ability / event rules
collect instance rules
choose destination
choose policy
produce TransitionPlan
```

## Resolver 表达能力

第一版不开放任意脚本化 resolver。动态目的地应先用两类能力表达：

```text
声明式条件表:
  按 actor、room、area、world、instance 等上下文检查有限谓词
  条件按顺序匹配，命中后返回目的地或候选目的地

内置 resolver 类型:
  框架提供常见动态移动能力
  内容只填写参数
```

普通出口仍然使用最简单的固定表达：

```text
StaticExit:
  direction: north
  to: village.road
```

声明式 DynamicExit 例子：

```text
DynamicExit enter mirror:
  rules:
    - when: actor.has_tag("dream_mark")
      to: dream_palace.entrance
    - when: world.time.is_night
      to: moon_corridor
    - otherwise:
      to: dusty_room
```

声明式条件表不能弱到只能检查“是否存在”。它应支持常见谓词，但仍然保持可检查、无副作用。

允许的谓词类型：

```text
存在 / 不存在:
  actor.has_tag("dream_mark")
  not actor.has_tag("cursed")

物品检查:
  actor.inventory.has_item("red_jade")
  not actor.inventory.has_item("iron_weapon")
  actor.inventory.count("ferry_ticket") >= 1

属性比较:
  actor.level >= 20
  actor.money >= fare
  actor.height_cm between 150 and 180
  actor.age_years < 16

枚举匹配:
  actor.faith in ["sea_god", "river_god"]
  actor.faction == "north_guard"
  area.civilization == "northern_clans"

状态检查:
  actor.state == spirit
  actor.combat_state == none
  instance.phase == completed

上下文检查:
  room.has_tag("recall_forbidden")
  area.danger_level >= D3
  world.time_of_day == night
  weather.kind == storm
```

不允许的内容：

```text
循环
任意函数调用
查询全世界
修改状态
生成物品
扣钱或发奖励
随机执行副作用
```

也就是说，条件表只能读上下文并选择目的地。扣费、清理状态、消耗物品、应用到达限制，必须放在 TransitionPolicy 或后续 PositionTransaction 流程中。

内置 resolver 例子：

```text
resolver: afterlife_by_culture
resolver: area_fallback
resolver: player_bound_home
resolver: ferry_route_current_stop
resolver: portal_by_key
resolver: route_selection
resolver: dungeon_safe_exit
```

不建议第一版支持：

```text
resolver_script: custom_mirror_destination(actor, room, world)
```

原因：

```text
脚本 resolver 难以静态检查
容易引入性能和 DoS 问题
调试复杂
可能破坏持久化和重启恢复语义
会过早把内容系统推向脚本运行时
```

原则：

> 90% 普通移动使用 StaticExit；常见特殊移动使用内置 resolver；复杂但可枚举的动态出口使用声明式条件表；任意脚本 resolver 后置。

例子：死亡。

```text
Trigger:
  player died

Context:
  actor faith = sea_god
  room danger_level = D2
  area afterlife = city_underworld
  faith afterlife override = drowned_harbor
  actor has soul_coin
  dungeon? none

Resolve:
  destination = drowned_harbor.entrance
  policy = D2 + sea_god afterlife cost model
  soul_coin opens reduced-cost gate

Execute:
  end combat
  clear temporary rewards
  move actor to drowned_harbor.entrance
  actor state = spirit
```

例子：回城符。

```text
Trigger:
  use recall_scroll

Context:
  room recall_forbidden? false
  actor bound_home = west_city.inn
  actor in combat? false
  scroll has charges

Resolve:
  destination = west_city.inn
  policy = consume scroll, clear selected temporary states

Execute:
  consume scroll
  move actor to west_city.inn
```

## 内容表达建议

内容作者可以按这个判断：

```text
如果它是从这个房间的某个方向或入口离开:
  写 Exit

如果 Exit 的目标固定:
  用 StaticExit

如果 Exit 的目标需要计算:
  用 DynamicExit

如果它由动作、物品、事件、状态触发:
  写 TransitionRule
```

也就是说：

```text
普通走路:
  Exit

同一个门去不同地方:
  DynamicExit

付费旅行、死亡、脱困、被捕、陷阱、梦境:
  TransitionRule
```

## 当前结论

```text
位置变更不归属于房间或人物之一。
房间、区域、人物、物品、能力、玩法实例都可以声明位置变更规则或修饰符。
所有规则最终生成统一的 TransitionRequest。
TransitionResolver 合成目的地和策略。
PositionTransaction 原子改变位置。
Exit 是 room-local TransitionRule。
StaticExit 是 FixedPlaceResolver + DefaultWalkPolicy。
DynamicExit 是 custom PlaceResolver + custom TransitionPolicy。
底层统一为 Transition；上层保留 Exit 作为普通地图的简化形式。
```

后续还需要讨论：

```text
DynamicExit 的 resolver 是否允许内容脚本，还是只允许声明式条件
多目标目的地如何让玩家选择
车费/船票这类 cost 是否走统一交易系统
Transition 是否允许移动多个 actor，例如船上所有乘客
Transition 失败时如何表达原因和候选方案
Transition 与隐藏出口、锁、门、钥匙系统如何合并
```
