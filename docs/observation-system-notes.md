# Observation 系统设计笔记

Observation 负责回答：某个观察者在某个上下文中能感知到什么。Presentation 负责把这些结构化观察结果渲染成文本。

命令和 Action 不直接拼接自然语言。`look`、`inventory`、`examine` 等动作只请求结构化观察结果；最终中文、纯文本、TUI 分区或其它客户端表现，都属于渲染层。

房间观察是 MUD 最重要的显示内容之一。大部分房间应该有独一无二的基础描述，但这并不意味着 `look` 结果应该是一整段硬编码文本。更好的模型是：

```text
Observation:
  生成结构化观察结果

Presentation:
  把观察结果按语言、视角、环境、隐藏条件和客户端能力渲染成文本
```

第一版命令路径：

```text
look / inventory / examine-like action
  -> build ObservationContext
  -> collect ObservationPayload / ObservationFacet
  -> emit structured PresentationEvent
  -> renderer turns payload into text or TUI view
```

禁止方向：

```text
look command concatenates final room text
get/drop/action handler writes Chinese sentence directly
tag provider returns already-stitched prose paragraph
```

允许方向：

```text
ObservationPayload:
  subject
  display fields
  visible entities
  visible exits
  facets
  message keys / structured fields
```

## 核心原则

```text
房间有独特基础描述。
上下文信息结构化生成。
通用氛围必须根据环境适配。
玩家可见物、出口、危险、天气、时间、光照都应是结构化字段。
Presentation 决定最终文本如何组合。
```

第一版可以有调试用 text renderer，但它消费的仍然是结构化 payload，而不是反过来要求规则层输出纯文本。

也就是说，房间不是只有一段 `description`，而是一个观察结果：

```text
RoomObservation:
  base_description
  atmosphere
  light
  time_context
  weather_context
  visible_entities
  visible_exits
  interactables
  danger_signals
  hidden_or_partial_signals
```

## 基础描述

大部分房间应有自己独一无二的基础描述。

```text
base_description:
  这是一条被车辙压实的土路，东边能看见镇墙，西边是一片低矮的松林。
```

基础描述表达房间的稳定身份：

```text
地形
建筑
方位感
长期存在的物件
文化与氛围
房间为何值得存在
```

基础描述不应承担所有动态信息。否则天气、时间、危险状态、可见实体、出口变化都要写进一段文本里，难以维护。

## 结构化上下文层

动态或半动态信息应结构化生成。

```text
time_context:
  morning / noon / dusk / night

weather_context:
  clear / rain / storm / snow / fog

light:
  bright / dim / dark / magical_darkness

visible_entities:
  players, NPCs, items, corpses, vehicles

visible_exits:
  north, gate, ferry, portal

danger_signals:
  blood smell, monster tracks, warning signs, hostile aura
```

Presentation 可以按需要组合：

```text
基础描述
时间/天气氛围
可见实体
明显出口
特殊提示
```

## 通用氛围需要环境适配

通用氛围不能简单套模板。

例如“晚霞染红了天空”适合开阔户外：

```text
你看到晚霞染红了天空。
```

但不适合：

```text
山洞里
地下室
密闭房间
明确写着万里无云但没有天空视野的场景
被魔法黑暗笼罩的区域
```

因此 atmosphere 规则必须检查环境上下文。

```text
AtmosphereContext:
  has_sky_view
  indoor / outdoor / cave / underwater / dream
  weather_exposed
  light_source
  area_climate
  room_tags
  explicit_overrides
```

例子：黄昏。

```text
outdoor + has_sky_view:
  晚霞染红了天空。

cave:
  洞外的光线渐渐暗了下来，洞中更显阴冷。

indoor with window:
  窗外透进一层昏红的余光。

sealed underground:
  你几乎感觉不到外面的时辰变化。

explicit no_sky_atmosphere:
  不显示天空相关氛围。
```

原则：

> 通用氛围不是全局文本插入，而是由 Observation 根据房间环境、区域、光照和覆盖规则生成的结构化提示。

## 房间 tag 对观察的影响

Room/Place tag 可以提供观察修饰。

```text
has_sky_view:
  可以显示天空、星月、晚霞

indoor:
  默认不直接显示天空天气，只显示间接影响

cave:
  天气影响弱，光照和回声更重要

underwater:
  使用水下观察规则

magical_darkness:
  限制可见实体和出口

safe_area:
  可显示安全感或人烟

danger_level:
  影响危险提示

no_generic_atmosphere:
  禁用通用氛围句

custom_atmosphere_profile:
  使用区域或房间专属氛围集合
```

这些 tag 不直接拼文本，而是参与 Observation 生成。

但 tag 不能默认暴露给玩家。内部规则 tag、任务 tag、掉落表 tag、反逃脱 tag 等如果自动进入观察结果，会泄露规则和内容设计。

原则：

> 只有显式声明 ObservationProvider / ObservationFacet 的强 tag，才可以贡献观察信息。

## ObservationFacet

强 tag 可以贡献结构化 ObservationFacet。

```text
ObservationFacet:
  kind
  source
  priority
  detail_level
  reveal_policy
  fields
  message_key?
```

kind 表示它是什么类型的观察信息：

```text
atmosphere
visibility_modifier
entity_listing
danger_signal
interactable_hint
exit_hint
condition_detail
resource_signal
tracking_signal
debug_signal
```

例子：`cave` tag 在黄昏时贡献一个氛围 facet。

```text
source: tag:cave
kind: atmosphere
priority: low
detail_level: detailed
message_key: atmosphere.cave_dusk
fields:
  outside_light: dimming
```

Presentation 可以渲染为：

```text
洞外的光线渐渐暗了下来，洞中更显阴冷。
```

例子：`danger_level D3` 可能贡献危险提示。

```text
source: area.danger_level
kind: danger_signal
priority: high
detail_level: brief
message_key: danger.signs_of_strong_monsters
reveal_policy: if_obvious_or_skilled
```

例子：内部任务 tag 不贡献观察信息。

```text
tag: quest_stage_17
ObservationFacet: none
```

## 观察分级与信息密度

Observation 必须支持简略和详细级别。否则玩家每次进入房间都会被大量通用信息刷屏，例如：

```text
太阳高悬于头顶。
你感到浑身发热。
酷暑时节。
空气干燥。
远方有蝉鸣。
```

这些信息可能在很大一片区域内都相同，默认反复显示会很烦。

推荐 detail level：

```text
brief:
  默认 look / 进入房间时显示
  只显示重要、变化明显、与行动相关的信息

standard:
  手动 look 时显示
  显示基础描述、明显实体、明显出口、重要氛围

detailed:
  look carefully / inspect / scan 时显示
  显示更多氛围、痕迹、弱提示、可交互线索

debug:
  管理员或 builder 查看
  显示规则来源、隐藏 tag、被过滤 facet
```

facet 进入最终输出前应按这些维度过滤：

```text
detail_level:
  当前观察模式是否要求这么详细

priority:
  是否足够重要

repetition:
  是否刚刚在相邻房间或短时间内显示过，以及是否到了提醒节奏

context_relevance:
  是否和当前行动有关

reveal_policy:
  观察者是否有能力或条件看到
```

默认进入房间或普通 look 不应展示所有低优先级通用氛围。只有当氛围异常、对行动有影响、第一次进入该区域、或者玩家使用详细观察时，才应该显示更多。

重复控制不应变成“显示一次后永远压制”。玩家在同一环境中长时间移动后，仍然需要被周期性提醒自己处在酷暑、暴雨、阴冷、死气弥漫等环境中。更合理的模型是：

```text
节奏化提醒:
  同一 atmosphere profile 隔一段时间或若干房间数提醒一次。

状态变化立即提醒:
  环境表述发生变化时立即显示。

机制影响立即提醒:
  环境开始影响行动、资源或危险判断时提升 priority。
```

ObservationFacet 可以携带 cadence 信息：

```text
cadence:
  profile_key: desert_heat
  state_key: dusk_hot
  show_on_enter_profile: true
  repeat_after_time: 10min
  repeat_after_rooms: 8
  show_on_change: true
  force_if_mechanical: true
```

Presentation/session 记住最近展示过的环境提示：

```text
last_shown_profile
last_shown_state
last_shown_time
last_shown_room_count
```

显示规则：

```text
如果 profile 第一次出现:
  显示

如果 state_key 变化:
  显示

如果机械影响开始或增强:
  显示

如果超过 repeat_after_time 或 repeat_after_rooms:
  显示

否则:
  暂时不显示
```

例子：荒漠酷暑。

```text
刚进入荒漠:
  夕阳下，热气仍从沙地里升起。

连续走几个房间:
  不重复刷酷热氛围。

过了一段时间:
  酷热仍然压在身上，汗水很快被风吹干。

黄昏变夜晚:
  沙地开始冷却，白日的灼热从空气中退去。

开始脱水:
  你感到喉咙发干，酷热正在消耗你的体力。
```

原则：

> 环境氛围需要 Presentation Cadence。默认不刷屏，但会周期性提醒；环境状态变化或产生机制影响时立即显示。

例子：酷暑。

```text
brief:
  如果酷暑没有直接影响行动，不显示。

standard:
  天气异常炎热时显示一句。

detailed:
  显示空气、汗水、地面热浪等更多细节。

mechanically relevant:
  如果酷暑正在造成 dehydration 或行动惩罚，提升 priority，brief 也显示。
```

原则：

> Observation 可以收集很多结构化 facet，但 Presentation 默认只显示重要内容。信息完整性保留在结构化 payload 中，文本输出必须控制密度。

## 可见性与观察者

Observation 必须依赖观察者。

同一房间，不同玩家可能看到不同内容：

```text
普通玩家:
  看见门、NPC、普通物品

夜视玩家:
  在黑暗中看见更多细节

有追踪技能的玩家:
  看见脚印、血迹、怪物活动痕迹

任务相关玩家:
  看见某个隐藏标记或梦境入口

管理员:
  看见 debug 信息
```

因此观察接口不应是：

```text
Room.Description() -> string
```

而应是：

```text
Observe(observer, target_place, observation_mode) -> ObservationPayload
```

## 观察模式

不同命令需要不同观察模式。

```text
look:
  标准房间观察

look carefully:
  更详细，可能显示更多 interactables 或线索

scan:
  偏向出口、危险和远处动静

track:
  偏向痕迹、足迹、气味

inspect object:
  观察单个对象

routes / exits:
  偏向出口和 Transition candidates
```

观察模式影响 ObservationPayload 的 detail level，而不是让每个命令各自拼文本。

## 出口与 Transition 的显示

出口和 Transition 也应进入结构化观察。

```text
visible_exits:
  - direction: north
    label: 北面的土路
    kind: static_exit
    obvious: true

  - command: board ferry
    label: 渡船
    kind: transition
    candidates: maybe hidden until routes

  - command: enter mirror
    label: 古镜
    kind: dynamic_exit
    reveal_policy: if_known
```

Presentation 可以选择简写：

```text
明显出口：北、东、渡船。
```

或详细显示：

```text
北面是一条通向镇外的土路。河边停着一艘渡船。
```

## 当前结论

```text
房间应保留独特基础描述。
动态上下文应结构化生成。
通用氛围必须根据 room/area/environment tag 适配。
Observation 依赖观察者、观察模式和当前上下文。
Presentation 负责把 ObservationPayload 渲染成最终文本。
Room.Description() -> string 不是合适的核心接口。
核心接口应类似 Observe(observer, target, mode) -> ObservationPayload。
```

后续还需要讨论：

```text
基础描述与动态段落如何排序
房间长描述和短描述是否分开
观察者技能如何影响隐藏信息
出口是否默认列出，还是融入描述
天气/时间氛围的 profile 如何配置
结构化客户端如何使用 ObservationPayload
```
