# 内容编译与 Runtime View 设计笔记

内容定义不应该直接作为运行时执行结构。内容层可以友好、声明式、适合编辑；运行时层必须适合快速查询、受控行为、热加载和错误检查。

核心原则：

> Tag 是内容语义；Runtime View 是执行结构。加载内容时应经过 parse、validate、normalize、compile，把强 tag / TagState 编译成适合运行时的 flags、typed slots、extension tag states、trigger index 和 provider lists。

## 为什么需要编译层

如果运行时每次都直接读取原始内容定义，会有几个问题：

```text
热点路径需要遍历 tag
内容错误到运行时才爆炸
trigger / requirement / resolver 名称可能拼错
tag 字段类型无法提前检查
热加载难以保证一致性
core slot 与 extension 的优化边界泄露给内容系统
```

编译层的目标是：

```text
内容作者写友好定义
编译器提前发现错误
运行时拿到高效结构
热加载通过 ContentSnapshot 切换
live state 不直接依赖原始文件结构
```

## Source Format

第一版内容源格式采用 CUE。Lua 只用于行为脚本，不用于定义实体或 tag。

```text
CUE:
  房间、区域、物品、NPC、出口、tag、display、trigger binding
  schema、默认值、约束、组合和静态校验

Lua:
  on_use / on_enter 等行为入口
  条件、选择器、概率和 effect request 组合
  不定义实体结构

Go ContentCompiler:
  通过 CUE 官方 Go API 读取、求值和校验 CUE
  加载 Lua script bindings
  生成规范化 RuntimeContent IR
  编译为 ContentSnapshot
```

CUE 集成使用 `cuelang.org/go` 官方 API，不通过 shell 调用 `cue export` 作为服务器运行路径。

```text
use:
  cuelang.org/go/cue/load
  cuelang.org/go/cue/cuecontext
  cue.Value.Validate(cue.Concrete(true))

do not use as runtime dependency:
  exec.Command("cue", "export", ...)
```

`cue` CLI 可以作为开发者人工调试工具，但不是服务端内容加载依赖。

不采用：

```text
YAML 作为第一版内容源
JSONC 作为主要内容源
数据库作为手写内容源
自制内容语言
Lua 定义实体或 tag
```

原则：

> CUE owns declarative content. Lua owns hot-loaded behavior. Go consumes normalized RuntimeContent, not arbitrary source files at runtime.

## Source to Runtime Boundary

ContentCompiler 是内容 source 到 runtime view 的唯一入口。World loop、command handler、relation system 和 presentation system 都不应直接读取 CUE AST、Lua 源文件路径或原始目录结构。

```text
CUE / Lua source
  -> CUE Go API + Lua loader
  -> ContentCompiler semantic validation
  -> RuntimeContent IR
  -> ContentSnapshot
  -> World runtime
```

第一版最小 RuntimeContent / ContentSnapshot 只需要表达：

```text
ContentSnapshot:
  version
  tag_classes
  entities
  spawns
  script_bindings
```

source 信息只用于错误报告、hot reload validation 和 developer diagnostics。runtime 不应依赖 CUE package layout、source file path、line/column、Lua module path 或目录 namespace 推导。

原则：

> ContentCompiler owns source semantics. ContentSnapshot is the only content view visible to world runtime.

## Implementation Order

实现顺序采用 `ContentSnapshot first, source loader second`。

第一阶段先手写一个开发用 `ContentSnapshot` fixture，跑通 world、relation、command、presentation 和 persistence。不要让 CUE API、Lua loader 或内容工具链阻塞第一批世界代码。

```text
phase 1:
  NewTutorialSnapshotForDev() -> ContentSnapshot
  hard-coded 2-3 rooms
  hard-coded 2-3 items
  hard-coded spawns

phase 2:
  LoadSnapshotFromCUE(path) -> ContentSnapshot
  uses CUE Go API
  same RuntimeContent / ContentSnapshot shape

phase 3:
  load Lua script bindings
  attach ScriptHandle to same ContentSnapshot model
```

约束：

```text
World depends only on ContentSnapshot.
World does not know whether snapshot came from fixture or CUE.
Fixture and CUE loader must produce the same runtime type.
```

手写 fixture 不是绕过内容系统，而是先固定 runtime 形状，让世界代码尽快可运行、可测试。

## Definition ID

第一版 `definition_id` 采用：

```text
kind.namespace.local_name
```

示例：

```text
area.tutorial
room.tutorial.start
room.tutorial.yard
room.tutorial.shed

item.tutorial.old_lantern
item.tutorial.oily_rag
item.tutorial.practice_sword

npc.tutorial.old_guard
```

规则：

```text
kind:
  area / room / item / npc / script / policy / resolver 等定义类型

namespace:
  通常是 area 或 shared namespace

local_name:
  该 namespace 内的稳定机器名
```

ID 约束：

```text
lowercase ascii only
dot separated
local_name 使用 snake_case
不允许空格
不允许中文
不允许路径分隔符
不混入玩家可见显示名
```

合法：

```text
room.tutorial.start
item.tutorial.old_lantern
npc.tutorial.old_guard
```

非法：

```text
room.Tutorial.Start
item.旧油灯
item.tutorial/old_lantern
tutorial.old lantern
```

玩家可见名称属于 display，不属于 ID：

```text
id: "item.tutorial.old_lantern"
display:
  name: "旧油灯"
```

原则：

> Definition IDs are stable machine references. Display names are localized player-facing text. Do not mix them.

第一版 CUE 内容里必须完整手写 `definition_id`，不由 compiler 隐式补全。

```text
required:
  id: "item.tutorial.old_lantern"

not first-version:
  id: "old_lantern" with implicit namespace/kind expansion
```

原因：第一版更需要清楚和可调试，不需要为了少写几个字符引入隐式规则。CUE 后续可以用 helper/template 降低重复，但进入 RuntimeContent 的 ID 必须总是完整 `kind.namespace.local_name`。

## Content Directory Layout

第一版内容目录按 area 聚合，而不是全局按类型堆大目录。MUD 内容作者通常是在制作一个区域；区域内的房间、物品、NPC、spawn 和脚本应该放在一起。

推荐结构：

```text
content/
  schema/
    content.cue
    room.cue
    item.cue
    npc.cue
    tag.cue

  shared/
    items.cue
    tags.cue
    scripts/
      common_lighting.lua
      common_doors.lua

  areas/
    tutorial/
      area.cue
      rooms.cue
      items.cue
      npcs.cue
      spawns.cue
      scripts/
        old_lantern.lua
        mirror_gate.lua
```

职责：

```text
schema/:
  全局 CUE schema、约束、默认值、类型定义

shared/:
  跨区域复用内容、通用 tags、通用脚本

areas/<area>/:
  区域自己的 area metadata、rooms、items、npcs、spawns、scripts
```

区域内文件拆分：

```text
area.cue:
  area id、namespace、起始房间、区域元数据、reset policy

rooms.cue:
  2-3 个房间、exits、display、room-level triggers

items.cue:
  2-3 个物品模板、tags、triggers

npcs.cue:
  NPC 模板；第一版可为空或省略

spawns.cue:
  初始房间放置哪些物品/NPC
```

原则：

> Area is the content work unit. Files inside an area split by concern. IDs remain explicit and complete even when directory namespace is obvious.

## EntityTemplate Shape

第一版只有一种实体模板：`EntityTemplate`。房间、物品、NPC 不是根本不同的模板类型，而是不同强 tags 组合形成的 entity。

文件仍然可以按作者习惯拆成 `rooms.cue`、`items.cue`、`npcs.cue`，但这些文件语义上都导出到统一的 `entities` map。

```text
rooms.cue -> entities
items.cue -> entities
npcs.cue  -> entities
```

`entities` map 的 key 和实体内部 `id` 都必须完整手写，并且必须一致。

```text
entities:
  "item.tutorial.old_lantern":
    id: "item.tutorial.old_lantern"
```

原因：map key 便于 CUE 合并、引用和查重；内部 `id` 让单个 entity block 自描述，错误报告也更清楚。ContentCompiler 必须校验 `entities` key 与 `id` 一致。

通用形状：

```text
entities:
  "<definition_id>":
    id: "<definition_id>"
    display: {...}
    tags: {...}
    state?: {...}
    reactions?: {...}
```

房间只是带 `place` / `observable` / `transition_source` 等 tags 的实体：

```text
entities:
  "room.tutorial.start":
    id: "room.tutorial.start"
    display:
      name: "练习场入口"
      look: "这里是练习场的入口。北边传来木剑碰撞的声音。"
    tags:
      place:
        kind: room
        allows_players: true
      observable: {}
      transition_source:
        exits:
          north:
            to: "room.tutorial.yard"
    reactions:
      on_enter:
        script: "areas/tutorial/scripts/start_room.lua:on_enter"
```

物品也是同一形状：

```text
entities:
  "item.tutorial.old_lantern":
    id: "item.tutorial.old_lantern"
    display:
      name: "旧油灯"
      look: "这盏油灯的铜壳已经发暗，里面似乎还有一点油。"
    tags:
      portable:
        weight: 1
      light_source:
        radius: 2
        lit_state: "lit"
    state:
      lit: false
    reactions:
      on_use:
        script: "areas/tutorial/scripts/old_lantern.lua:on_use"
```

NPC 也是同一形状：

```text
entities:
  "npc.tutorial.old_guard":
    id: "npc.tutorial.old_guard"
    display:
      name: "老守卫"
      look: "他拄着木剑，似乎正在打盹。"
    tags:
      actor: {}
      npc: {}
      observable: {}
```

原则：

> Room, item, and NPC are authoring conveniences, not separate semantic roots. Everything compiles to EntityTemplate with strong tags.

## SpawnDefinition and Relations

`EntityTemplate` 不保存固定位置。模板只说明“这个东西是什么”，不说明“这个具体实例现在在哪里”。

需要分开三层：

```text
EntityTemplate:
  定义旧油灯是什么

SpawnDefinition:
  定义起始房间初始化或 reset 时应该生成一盏旧油灯

LiveRelation:
  定义这盏具体旧油灯实例现在在哪里
```

不要在物品模板里写：

```text
item.tutorial.old_lantern:
  location: room.tutorial.start
```

推荐在 `spawns.cue` 中写：

```text
spawns:
  "spawn.tutorial.start_lantern":
    id: "spawn.tutorial.start_lantern"
    entity: "item.tutorial.old_lantern"
    target: "room.tutorial.start"
    relation: "inside"
    lifecycle: "reset_managed"
```

第一版最小 relation 只需要 `inside`：

```text
player inside room
item inside room
item inside player inventory
```

这已经能验证：

```text
房间包含玩家
房间包含地上物
玩家背包包含物品
get/drop 改变 relation
重连/重启保存 relation
```

后续再扩展：

```text
equipped_in
attached_to
held_by
contained_in
mounted_on
```

第一版每个 `SpawnDefinition` 只生成一个实例，不支持 `count` 或 spawn group。需要两个实例就写两个 spawn。

```text
spawns:
  "spawn.tutorial.coin_1":
    id: "spawn.tutorial.coin_1"
    entity: "item.tutorial.coin"
    target: "room.tutorial.start"
    relation: "inside"
    lifecycle: "reset_managed"

  "spawn.tutorial.coin_2":
    id: "spawn.tutorial.coin_2"
    entity: "item.tutorial.coin"
    target: "room.tutorial.start"
    relation: "inside"
    lifecycle: "reset_managed"
```

第一版不做：

```text
count: 3
spawn group
weighted spawn
conditional spawn
recursive container spawn
drop_tables
loot tables
```

DropTable 不进入第一版 ContentCompiler 范围。第一版只编译 `entities`、`spawns`、script bindings 和强 tag schema。掉落表等到战斗、宝箱、采集或副本奖励有真实触发点时再设计。

原则：

> Templates define what things are. Spawns define how initial/reset instances appear. LiveRelation defines where concrete instances are now. First-version spawns are one-definition, one-instance.

## 流程

推荐加载流程：

```text
content files
  -> parse
  -> validate
  -> normalize
  -> compile
  -> ContentSnapshot
  -> RuntimeEntity templates
```

第一版加载校验只覆盖会影响运行时正确性的边界：

```text
所有 definition_id 唯一
entities map key == entity.id
所有 spawn.entity 指向存在的 entity
所有 spawn.target 指向存在且可作为 inside 目标的 entity
spawn.relation 第一版只能是 inside
所有 entity tags 都是已注册 TagClass
tag 属性符合 schema
所有 reaction hook 已注册
所有 script path / entry 可加载
新 ContentSnapshot 与当前 live references 兼容
```

第一版不做复杂 content migration、复杂 include/template system、复杂 resolver/policy registry、DropTable 校验、跨 shard / 多世界 content version 或编辑器级错误恢复。

### Parse

解析内容文件，保留原始位置用于错误报告。

```text
source file
line / column
definition id
raw tags
raw triggers
raw requirements
```

### Validate

尽量在启动或热加载时发现内容错误。

应检查：

```text
definition_id 唯一
未知 tag
未知 trigger kind
未知 requirement kind
未知 resolver / policy
字段类型错误
必填字段缺失
互斥能力冲突
引用目标不存在
数值范围非法
循环引用或不允许的递归
```

例子：

```text
Exit requires target place exists.
Container capacity must be positive.
DangerLevel must be D0-D4.
Transition resolver must be registered.
ObservationFacet reveal_policy must be known.
```

### Normalize

把多种友好写法归一成统一中间形式。

例如，marker tag 短写：

```text
tags: [portable, flammable]
```

可以归一为：

```text
tags:
  portable: {}
  flammable: {}
```

短写出口：

```text
north: room_road
```

归一为：

```text
transition:
  trigger: command("north")
  resolver: fixed_place(room_road)
  policy: default_walk
```

### Compile

把 normalized definition 编译成 Runtime View。

```text
RuntimeEntityTemplate:
  definition_id
  flags
  core_slots
  extension_tag_states
  trigger_index
  script_bindings
  observation_providers
  transition_providers
  effect_providers
  display metadata
```

## 示例

内容定义：

```text
lantern:
  tags:
    portable: {}
    flammable: {}
    light_source:
      radius: 2
    fuel_consumer:
      fuel: oil
      duration: 30min
    durability:
      max: 50
  triggers:
    on_take_damage:
      - when: damage_type == fire
        effect: ignite
```

编译后：

```text
RuntimeEntityTemplate:
  flags:
    Portable
    Flammable
  core_slots:
    durability?    # 如果 durability 已晋升 core
  extension_tag_states:
    light_source
    fuel_consumer
  trigger_index:
    on_take_damage -> [ignite_rule]
  observation_providers:
    light_source_provider
    flammable_provider
  effect_providers:
    fuel_consumer_provider
```

运行时攻击、观察、移动、tick 都不需要遍历原始 tag 列表。

## TagClass 与 Runtime Layout

内容定义不应关心某能力在 core slot 还是 extension map。

内容可以始终写：

```text
tags:
  durability:
    max: 100
```

编译器根据当前 runtime schema 决定放哪里：

```text
if durability is core:
  template.core_slots.durability = DurabilityTagState(...)
else:
  template.extension_tag_states["durability"] = DurabilityTagState(...)
```

内容层的 tag 是类式语义对象，而不是布尔标签。

```text
TagClass:
  name
  schema
  methods / behavior API
  lua method bindings
  Go primitive / provider references
  observation providers
  transition providers
  effect providers
  conflict rules

TagInstance / TagState:
  entity 上该 tag 的具体属性值
```

例子：

```text
tags:
  place:
    kind: room
  light_source:
    radius: 2
  portable: {}
```

这里 `place`、`light_source`、`portable` 都是 TagClass 的实例化。房间不是特殊根类型，而是具备 `place`、`observable`、`transition_source` 等 tags 的实体。

TagClass 方法不是裸 Lua 函数，也不是裸 Go 方法。它们可以有两种绑定：

```text
Lua method / reaction binding:
  内容可热加载
  用 Lua 控制流组合 condition / selector / effect primitive
  返回 effect requests

Go primitive / provider binding:
  引擎核心能力
  高频路径或强事务边界
  直接由 Go 实现，不经过 Lua dispatch
```

编译后 Runtime View 可以包含：

```text
TagRuntimeBinding:
  tag_name
  tag_state_layout
  lua_method_bindings
  go_provider_refs
  effect_capabilities
  trigger_hooks
```

例子：

```text
portable:
  lua methods:
    on_pickup?
    on_drop?
  Go primitives/providers:
    Inventory.CanPickUp
    PositionTransaction.Move

place:
  Go providers:
    Place.CanContain
    Transition.AvailableExits
    Observation.Scope
  Lua reactions:
    on_enter?
    on_leave?
```

第一版原则：Tag 方法可以通过 Lua 组合 Go primitive，但核心执行仍属于 Go。极底层、高频、完全核心的能力可以直接注册为 Go provider，不需要绕 Lua。

外部访问仍然通过稳定 accessor：

```text
item.Durability()
```

原则：

> 内容语义不随运行时优化边界变化。core/extension 的迁移由 compiler 和 accessor 隔离。

## Trigger Index

trigger 不应在运行时扫描全部 tag 查找。

编译器应按 trigger kind 建索引：

```text
trigger_index:
  on_attack_hit -> [bleed_rule, twist_space_rule]
  on_enter_room -> [trap_rule]
  on_observe -> [glow_hint_rule]
  on_death -> [soul_anchor_rule]
```

运行时只查当前事件对应列表：

```text
weapon.Triggers(OnAttackHit)
room.Triggers(OnEnterRoom)
actor.Triggers(OnDeath)
```

## Script Bindings

脚本不是 Tag / Component 的替代品。脚本用于需要免 Go 编译、频繁迭代的行为边界，例如特殊房间机关、物品使用反应、NPC 对话、任务步骤、少量自定义 resolver 或 trigger reaction。

内容定义可以引用脚本入口，但这些引用仍应经过 ContentCompiler 校验：

```text
triggers:
  on_use:
    script: scripts/items/old_lantern.lua:on_use

transition:
  resolver_script: scripts/rooms/mirror_gate.lua:resolve
```

脚本可以有两种形态：

```text
structured behavior table:
  Lua 返回结构化 behavior definition
  适合简单行为
  编译期更容易校验

function entry:
  Lua 导出符合 hook kind 的函数
  运行时读取上下文、调用 condition / selector helper
  返回 effect request list
  适合需要控制流的复杂行为
```

示例：结构化行为。

```lua
return behavior {
  hook = "on_use",
  conditions = {
    has_tag(self(), "light_source"),
  },
  effects = {
    emit("你点亮了灯。"),
    set_state(self(), "lit", true),
  },
}
```

示例：函数入口。

```lua
function on_use(ctx)
  if not conditions.has_tag(ctx.actor, "can_read_runes") then
    return { effects.emit("你看不懂这些符文。") }
  end

  if conditions.chance(ctx, 30) then
    return {
      effects.emit("石门缓缓打开。"),
      effects.set_state(ctx.self, "open", true),
    }
  end

  return { effects.emit("符文闪了一下，然后归于沉寂。") }
end
```

编译阶段应检查：

```text
脚本文件存在
入口函数存在
入口签名符合该 hook kind
结构化 behavior table 的字段合法
函数入口声明的返回类型是 effect request list
脚本声明的权限/API capability 合法
脚本依赖的 definition_id / resolver / effect 名称存在
脚本可被 sandbox 加载
```

编译后 Runtime View 不保存任意原始脚本文本查询，而保存受控绑定：

```text
script_bindings:
  on_use -> ScriptHandle(old_lantern:on_use)
  transition_resolver -> ScriptHandle(mirror_gate:resolve)
```

脚本和声明式内容一起进入 `ContentSnapshot`。热加载时构建新 snapshot，脚本加载、签名检查和权限检查失败则整次 reload 失败，旧 snapshot 继续服务。

### Content and Script Layout

声明式内容是主入口；Lua 脚本是被内容引用的行为模块。脚本不反向定义实体，不声明 tag，不成为隐藏的内容来源。

推荐目录组织：

```text
content/
  areas/
    tutorial/
      rooms.cue
      items.cue
      npcs.cue
      scripts/
        old_lantern.lua
        mirror_gate.lua

  shared/
    scripts/
      common_lighting.lua
      common_doors.lua
```

规则：

```text
area-local scripts:
  区域专属房间机关
  区域专属物品行为
  区域专属 NPC 对话或任务步骤

shared scripts:
  多个区域复用的通用行为
  通用门、灯、容器、提示、基础机关

declarative content:
  定义 entity、tag、relation、trigger binding
  引用 script entry
```

示例：

```text
item.old_lantern:
  tags:
    portable: {}
    light_source:
      radius: 2
  triggers:
    on_use:
      script: areas/tutorial/scripts/old_lantern.lua:on_use
```

热加载单位仍然是完整 `ContentSnapshot`：

```text
load declarative content
load area-local scripts
load shared scripts
validate bindings
build new ContentSnapshot
swap if successful
```

原则：

> Content owns structure; scripts provide behavior. Area-local scripts keep special content nearby; shared scripts prevent copy-paste common behavior.

### Behavior Primitive Registry

脚本层只允许引用引擎注册过的行为 primitive。ContentCompiler 应把这些 primitive 当成有限词汇表校验，而不是允许脚本任意修改世界。

```text
BehaviorPrimitiveRegistry:
  hooks:
    on_use
    on_enter
    on_pickup

  conditions:
    has_tag
    state_equals
    chance

  selectors:
    actor
    self
    room

  effects:
    emit
    set_state
    move_item
```

编译阶段应检查脚本只引用已注册 primitive：

```text
unknown hook:
  validation error

unknown condition / selector / effect:
  validation error

wrong argument shape:
  validation error

effect not allowed by hook capability:
  validation error
```

函数入口的完整 effect list 可能只能在运行时根据上下文产生，因此需要两层验证：

```text
load-time validation:
  脚本入口、签名、capability、可引用 primitive、显式声明依赖

run-time validation:
  函数返回值必须是 effect request list
  每个 effect request 的 primitive 已注册
  参数形状合法
  目标仍有效
  当前 hook 允许该 effect
  当前 actor / item / room 有权限触发该 effect
```

结构化 behavior table 可以在 load-time 校验更多内容；函数入口保留 Lua 控制流，但必须接受 run-time effect validation。

### Script Execution Budget

Lua 函数可以使用控制流，因此每次 hook 调用必须有执行预算。预算是 sandbox 的一部分，用来防止死循环、递归爆炸、超长查询和超大量 effect request。

```text
ScriptExecutionBudget:
  instruction_limit
  wall_time_limit
  effect_count_limit
  query_count_limit
  stack / recursion limit
```

第一版采用按 hook kind 固定预算，不允许内容脚本自行声明更高预算。

```text
on_use:
  small budget
  few queries
  few effects

on_enter:
  small budget
  few queries
  few effects

future scheduled_event:
  medium budget

future world_event:
  medium or large budget, admin-reviewed
```

后续可以引入声明式 budget tier，但需要权限或内容审查。

```text
budget: small
budget: medium
budget: large  # wizard/admin reviewed only
```

超预算处理：

```text
instruction limit exceeded:
  script failure

wall time exceeded:
  script failure

too many effect requests:
  script failure

too many selector/query calls:
  script failure
```

脚本超预算和脚本报错一样处理：不提交任何 effect，生成 `ScriptErrorEvent`，普通玩家只看到错误码，测试用户/巫师按权限看到摘要。

原则：

> Lua 可以有控制流，但每次 hook 调用必须有预算。内容脚本可以灵活组合行为，不能无限占用 world loop。

### Script RNG

Lua 脚本不应使用自己的全局随机数。

```text
forbidden:
  math.random()
  script-local RNG seeded by wall clock
  looping random attempts until success
```

随机由 Go 在每次事件或 hook 调用时创建局部 RNG context。脚本只能通过注册 primitive 消耗随机，例如 `conditions.chance(ctx, percent)`。

```text
RngContext:
  domain
  event_id
  script_id
  hook
  actor_id?
  entity_id?
  seed / derivation input
  draw_count
```

随机不应是全世界共享的一个巨大序列。全局序列会让系统互相污染：某个脚本多消耗一次随机数，可能改变另一个玩家的掉落、战斗或机关结果。

推荐模型：

```text
WorldRngRoot:
  server/world root seed or secret

RngDomain:
  script
  combat
  loot
  spawn
  dungeon
  cosmetic

Hook-local RNG:
  derive(root, domain, event_id, script_id, hook, actor/entity context)
```

这样脚本随机具备隔离性和可追踪性：

```text
script 多一次 chance:
  不影响 loot RNG
  不影响 combat RNG
  不推进全局随机序列

debug trace:
  draw_index
  probability
  roll/result
  script_id
  hook
```

第一版只需要一个随机 primitive：

```text
conditions.chance(ctx, percent)
```

后续再考虑：

```text
rng.choose(ctx, candidates)
rng.weighted_choice(ctx, weighted_table)
rng.roll(ctx, dice_expression)
```

随机 draw 也计入 hook execution budget。超出 `rng_draw_limit` 或通过非法 API 取随机，视为脚本失败。

原则：

> Randomness is event-local, domain-separated, Go-owned, budgeted, and traceable. Lua can ask for chance; Lua cannot own randomness.

### Script Runtime Errors

脚本运行失败或 effect validation 失败时，不能提交半完成的世界变化。

```text
Lua execute
  -> collect effect requests
  -> validate full list
  -> build WorldMutation
  -> apply atomically
```

任意一步失败：

```text
不提交任何 effect
不修改 relation graph
不写入玩家利益状态
不触发后续 reaction
生成 ScriptErrorEvent
```

`ScriptErrorEvent` 应包含完整内部信息和一个可报告的短错误码。

```text
ScriptErrorEvent:
  error_code
  content_snapshot_version
  script_module
  script_entry
  hook
  actor_id?
  entity_id?
  place_id?
  internal_error
  stack_trace?
  effect_request?
```

错误码用于玩家报告，不应泄露脚本源码、路径细节、内部状态或堆栈。

```text
normal player:
  你感觉有什么东西出错了。错误码 MUD-8F3A2C。若你看到这个，请尽快联系巫师。

test user / wizard / developer:
  MUD-8F3A2C script error in scripts/items/old_lantern.lua:on_use
  hook=on_use snapshot=content-42 reason=unknown effect move_entity
```

显示策略由 actor/session 权限决定，而不是只由 server mode 决定。

```text
normal player:
  show reportable error_code + contact-wizard message
  hide raw error text

test user:
  show error_code + script id + hook + short reason

wizard / developer:
  show expanded diagnostic if permission allows

server log / monitoring:
  store full internal error and stack trace
```

原则：

> 普通玩家拿到可报告错误码；测试用户和巫师拿到可定位摘要；完整错误只进入内部日志。错误反馈不能泄露实现细节，也不能让脚本失败产生半完成世界状态。

`move_item` 是特殊但必要的 relation mutation primitive。它必须通过 Go 的 `PositionTransaction` 或等价服务执行，不能由 Lua 直接改 relation graph。

```text
move_item effect request:
  item
  from
  to
  reason

Go validation:
  item 当前确实在 from
  to 是当前 hook 允许的目标
  actor 有权限移动该 item
  item 没有绑定/重量/容量/状态限制阻止移动
  移动不会破坏唯一位置约束

Go apply:
  PositionTransaction.Move(item, from, to)
  标记 observation dirty
  触发必要 reaction
  更新持久化边界
```

第一版 `move_item` 应限制范围：

```text
allowed:
  room ground -> actor inventory
  actor inventory -> room ground

not yet:
  equipment slot -> room ground
  arbitrary container moves
  other player's inventory
  remote room moves
  batch moves
  create / destroy
```

未来像“给武器涂油后可点燃但可能从手中滑落”这类行为，可以扩展为 `equipment slot -> room ground`，但仍然走同一个 `move_item` validation/apply 管线。

这形成两层迭代速度：

```text
内容组合变化:
  使用已有 hook / condition / selector / effect
  只需要重新编译 ContentSnapshot
  不需要重新编译 Go

引擎词汇表扩展:
  新增 hook / condition / selector / effect primitive
  需要修改 Go registry 和执行器
  需要重新编译 Go
```

这个取舍是可接受的。日常内容开发主要是在既有词汇表内组合新行为；随着游戏开发推进，新增 primitive 的频率应该越来越低。

原则：

> 脚本提供免 Go 编译的内容组合；BehaviorPrimitiveRegistry 定义需要 Go 编译的引擎词汇表。Lua 可以使用控制流，但只能返回 effect requests；运行时不能让脚本绕过事务、持久化和权限模型。

## Provider Lists

Observation、Transition、Effect 都可以通过 provider list 避免遍历所有 tag。

```text
observation_providers:
  sky_view_provider
  danger_signal_provider
  light_source_provider

transition_providers:
  static_exit_provider
  dynamic_exit_provider
  fallback_provider

effect_providers:
  flammable_provider
  poison_provider
  durability_provider
```

provider 只在相关系统需要时被调用，且范围应是局部集合。

## ContentSnapshot

一次成功编译的内容集形成 ContentSnapshot。

```text
ContentSnapshot:
  version
  definitions
  runtime_templates
  script_modules
  resolver registry
  policy registry
  validation report
```

热加载流程：

```text
compile new ContentSnapshot
if validation fails:
  keep old snapshot
  report errors

if validation passes:
  atomically switch active snapshot
  new spawns use new templates
  live state keeps identity and player-interest state
```

LiveState 不应该直接等于内容定义。

```text
LiveEntity:
  entity_id
  definition_id
  content_version_seen
  mutable_state
  player_interest_state?
```

## 热加载与 Live State

热加载不等于世界 reset。

```text
内容定义改变:
  新生成实体使用新模板
  已存在 live entity 保留 entity_id / definition_id / mutable_state
  行为与展示默认读取新的 active ContentSnapshot

玩家利益状态:
  不因为内容热加载丢失

临时 live world:
  可按 reset / reload policy 处理
```

第一版采用简单、安全的规则：

```text
LiveEntity:
  持有 definition_id
  不复制完整 RuntimeEntityTemplate

active ContentSnapshot:
  提供当前展示、TagClass 定义、script binding、provider lists

LiveState:
  保存实例事实和玩家利益
  例如位置、背包归属、耐久、绑定、词条、当前 open/lit/charged state
```

因此：

```text
描述文本修改:
  下次 look 使用新描述

script on_use 修改:
  下次 on_use 使用新 script binding

light_source radius 修改:
  下次相关查询使用新模板值

玩家背包里的已确认物品:
  不因为模板 reload 消失或重置
```

热加载前必须校验当前 live world 对新 snapshot 的引用仍然兼容：

```text
all live definition_id exists
player current place exists or has explicit migration/fallback policy
required tag schema compatible
script bindings referenced by live entities are valid
resolver / policy / provider names still registered
mutable_state keys required by new runtime template are compatible
```

第一版不做自动迁移。如果新 snapshot 与 live references 不兼容：

```text
reject new ContentSnapshot
keep old snapshot active
report validation errors
do not partially reload
```

后续再考虑：

```text
content migration
deprecated definition alias
live entity quarantine
admin repair command
explicit place migration rule
```

如果内容变更破坏 live entity，例如删除玩家所在房间：

```text
第一版默认拒绝热加载
除非新 snapshot 提供显式迁移或 fallback policy
```

## 编译错误与报告

错误报告必须面向内容作者。

```text
file: areas/river.yaml
line: 42
definition: east_ferry_dock
error: unknown transition resolver "ferry_by_moon"
hint: registered resolvers are fixed_place, ferry_route_current_stop, area_fallback
```

不要只报运行时 panic。

## 当前结论

```text
内容定义是友好声明式输入。
Runtime View 是高效执行结构。
加载时必须 parse / validate / normalize / compile。
强 tag / TagState 编译成 flags、typed slots、extension tag states、trigger index 和 provider lists。
行为脚本可以作为 script_bindings / script_modules 进入 ContentSnapshot，但不能取代 tag 的声明式语义。
core slot 与 extension 的区别由 compiler 隔离，不泄露给内容语义。
ContentSnapshot 是热加载单位。
LiveState 引用 definition_id 和 snapshot version，但保存自己的 mutable/player-interest state。
已存在 live entity 默认从 active ContentSnapshot 读取行为和展示；实例事实不被模板覆盖。
新 snapshot 与 live references 不兼容时，第一版拒绝热加载，不自动迁移。
```

后续还需要讨论：

```text
definition_id 命名规范
ContentSnapshot 版本与 live entity migration
编译错误如何在 TUI/admin 界面显示
resolver/policy registry 如何注册
脚本 sandbox、权限、API 稳定性和热加载错误报告
```
