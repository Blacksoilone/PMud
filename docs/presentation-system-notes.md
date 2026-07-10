# Presentation 系统设计笔记

Presentation 负责把系统产生的结构化事实渲染成玩家看到的文本。框架层、规则层、Action、Transition、Death、Requirement 都不应该直接拼接自然语言。

所有会产生玩家可见内容的动作，都只请求或产出结构化 payload。文本拼接、排序、换行、分区、颜色、强调、TUI 布局和纯文本 fallback 都属于 renderer。

核心原则：

> 所有玩家可见文本都应从结构化字段解析出来。系统产出结构化事实、原因、候选和事件；Presentation/i18n 层负责选择模板、代入字段、调整视角并渲染文本。

第一版允许实现一个简单 text renderer，用于调试和最小可玩闭环。但它必须消费 `PresentationEvent` / `ObservationPayload`，不能让命令层直接返回最终中文段落。

```text
command/action:
  returns structured facts

presentation:
  wraps facts as PresentationEvent

renderer:
  concatenates text / builds TUI regions
```

## 为什么要结构化输出

MUD 看似只有文本，但如果到处硬编码文本，会很快遇到问题：

```text
无法国际化
无法根据观察者视角改变描述
无法给客户端结构化输出
无法统一失败原因和提示
无法避免隐藏条件剧透
无法调试为什么某个动作失败
同一事件在 actor、target、observer 眼里难以区分
```

因此，规则系统不应该返回：

```text
"你没有船票。"
"守卫只允许成年人进入。"
"你正在战斗中，不能回城。"
```

而应返回结构化 payload：

```text
PresentationEvent:
  type: transition_blocked
  message_key: missing_required_item
  actor: player_123
  item: ferry_ticket
  source: inventory
  hint_key: buy_ticket_at_dock
```

Presentation 再根据接收者、语言、风格和可见信息渲染：

```text
你没有船票。
你可以在码头售票人那里购买船票。
```

## PresentationEvent

推荐所有可见反馈都统一为 PresentationEvent 或类似结构。

```text
PresentationEvent:
  event_type
  audience
  message_key
  severity
  entities
  values
  candidates?
  blockers?
  reveal_policy?
  debug?
```

PresentationEvent 还应携带客户端渲染所需的语义通道和显示建议。

```text
PresentationEvent:
  channel
  priority
  suggested_region?
  full_text_fallback?
```

channel 可以是：

```text
room:
  房间基础描述、观察结果

ambient:
  天气、时间、环境氛围

status:
  血量、气力、饥渴、死亡状态、escape 限制

combat:
  战斗事件

system:
  错误、警告、连接状态

chat:
  say / tell / channel

transition:
  移动、渡船、传送、死亡进入死后世界

prompt:
  输入提示、候选、确认请求
```

服务器可以给 `suggested_region`，但不应硬编码客户端布局。不同客户端可以用同一事件流做不同渲染。

```text
纯文本客户端:
  把事件按顺序渲染到滚动文本流

增强 TUI 客户端:
  房间区、状态区、日志区、输入区、候选/确认区分开显示

未来 Web / GUI 客户端:
  把同一结构化 payload 渲染成卡片、列表、状态栏或按钮
```

因此，纯文本输出是 renderer 的一种实现，不是规则层接口。

传统 MUD 通常以任意终端 telnet 兼容为核心目标。本项目长期不以 telnet 或任意终端兼容为核心体验，而是保留固定结构化客户端 / TUI 客户端方向。

但第一版手写实现优先恢复代码能力，可以使用 TCP line transport + text renderer 作为调试和最小可玩入口。纯文本输出是结构化 payload 的 renderer，不是规则层接口，也不是长期客户端能力的上限。

长期结构化客户端倾向使用 WebSocket + JSON 事件流。原因是固定客户端天然可以维护双向结构化事件流；WebSocket 也方便未来浏览器客户端、TUI 客户端和调试工具共用同一协议。

长期消息编码先使用 JSON：

```text
client -> server:
  JSON command envelope

server -> client:
  JSON PresentationEvent stream
```

命令和事件未来可以演进为二进制协议，但不应在协议形态还频繁变化时过早引入。JSON 的优势是可读、易调试、Go 端和客户端都容易实现，日志和回放也更方便。等消息 schema 稳定后，再考虑 protobuf、msgpack 或自定义二进制帧。

原则：

> 固定结构化客户端 / TUI 是长期核心体验。服务器输出结构化 PresentationEvent，客户端决定如何分区显示。第一版可以用 TCP line + text renderer 快速跑通，但不反向约束核心协议。

> 长期传输倾向 WebSocket + JSON 事件流。二进制命令/事件协议后置到 schema 稳定之后。

## 客户端纯渲染与资源包边界

长期客户端应是纯渲染层。客户端可以保存整个游戏的显示资源包，但不保存世界关系、规则、隐藏条件或当前世界状态。

```text
client owns:
  text resources
  i18n catalogs
  style/theme choices
  layout and TUI rendering
  input editing and local aliases

server owns:
  world graph
  tags and entity composition
  rules and requirements
  random rolls and combat resolution
  hidden exits and reveal conditions
  player/NPC/item state
  observation filtering
```

这意味着客户端资源包可以包含未来可能出现的文本：

```text
room.hidden_cave.name = 隐藏洞窟
room.hidden_cave.description = 墙后传来潮湿的风。
item.old_lantern.name = 旧油灯
npc.old_guard.hint_01 = 老守卫低声说：夜里别靠近井边。
```

但这些文本本身不携带相互关系。玩家即使解包资源，也只得到散页词典：知道某个 key 对应什么显示内容，不知道它什么时候出现、由哪个 tag/provider 暴露、从哪里进入、需要什么条件、会触发什么规则。

客户端不应收到或内置：

```text
room.hidden_cave.enter_from = room.tutorial.start
room.hidden_cave.requires = item.rusty_key
drop.goblin_king.legendary_sword.probability = 0.001
tag.secret_wall.providers = hidden_cave_exit
```

这些属于服务端世界事实和规则。服务端只在玩家当前可观察时，向客户端发送可观察对象的 id、资源 key 和必要状态 hint。

```text
event=room_observed
room=room.tutorial.start
name_key=room.tutorial.start.name
description_key=room.tutorial.start.description
exits=north
items=item.old_lantern
```

客户端根据 `*_key` 查本地资源表，再选择如何展示。改变语言、字体、边框、地图样式、列表布局或无障碍呈现，都只改客户端和资源包，不要求服务端准备多套输出。

原则：

> Client owns text resources and rendering. Server owns world graph, rules, state, and observation filtering.

> 客户端资源泄露的上限是剧透文本；真正需要保护的是未观察世界事实不要下发。

## 多 facet 描述组合

描述资源不能假设“一字段对应一句描述”。很多高质量 MUD 文本来自多个可观察 facet 的融合，而不是机械列举每个状态。

错误方向：

```text
weather=sunny  -> 万里无云。
time=noon      -> 太阳挂在正上方。
temperature=hot -> 你感到浑身燥热。
```

直接拼接后会变成：

```text
万里无云。太阳挂在正上方。你感到浑身燥热。
```

更好的表达是由多个 facet 命中一个融合描述：

```text
facets = [weather.sunny, time.noon, temperature.hot]
  -> atmosphere.sunny_noon_hot
  -> 万里无云，正午的阳光晒得地面发白。
```

长期模型：服务端发送当前玩家可观察到的 facets；客户端不自由创作文本，而是使用 content compiler 生成的组合索引选择 prose key。

```text
server event:
  weather=weather.sunny
  time=time.noon
  temperature=temperature.hot

client compiled lookup:
  [weather.sunny, time.noon, temperature.hot]
    -> atmosphere.sunny_noon_hot

client catalog:
  atmosphere.sunny_noon_hot = 万里无云，正午的阳光晒得地面发白。
```

如果最具体组合不存在，客户端按编译好的降级规则查较小组合或 fragment。这个过程仍然是查表，不是运行时 NLG。

```text
1. [weather.sunny, time.noon, temperature.hot]
2. [weather.sunny, time.noon]
3. [weather.sunny, temperature.hot]
4. [time.noon, temperature.hot]
5. single-facet fragments
```

同样原则适用于物品观察深度：

```text
server event:
  item=item.iron_sword
  view=detail.brief
  condition=condition.rusty

client compiled lookup:
  [item.iron_sword, detail.brief, condition.rusty]
    -> item.iron_sword.description.brief_rusty
```

服务端仍负责观察裁剪，例如玩家只能看到 `detail.brief` 而不是 `detail.full`。客户端负责把这些可观察 facet 通过资源包表达成自然文本。

原则：

> Server sends observable facets. Client selects prose through compiler-generated lookup tables. Client does not invent prose at runtime.

> 会影响交互、布局或玩家选择的事实应作为字段发送；多个显示 facet 可以共同选择一个描述资源；不要为了方便同时发送等价的最终 key 和组成 facets。

## TUI 输入原则

固定 TUI 客户端不等于按钮游戏。第一版不依赖鼠标事件；所有操作原则上都应能通过文字命令完成。

```text
必须支持:
  文本命令输入
  数字候选选择
  Enter 确认
  Esc 取消
  Tab / 方向键移动焦点或选择候选

暂不支持:
  鼠标点击作为核心输入
```

快捷键只是效率层，不是唯一入口。

例如玩家可以直接输入：

```text
ride cart to luoyang
```

也可以先输入：

```text
ride cart
```

客户端显示候选：

```text
[1] 洛阳驿站  50 文
[2] 开封驿站  80 文
[3] 扬州码头 120 文
```

然后玩家按 `1`，客户端发送结构化选择：

```text
kind: candidate_select
request_id: travel_123
candidate_id: luoyang_station
```

这两个入口最终表达同一游戏意图。

原则：

> 命令完整覆盖能力；Tab、数字、方向键、确认/取消只提供快捷操作。客户端快捷键最终也应转成结构化 command envelope，而不是绕过命令/规则系统。

### event_type

表示事件类型。

```text
action_success
action_failed
transition_success
transition_blocked
death_result
afterlife_arrival
ambiguity
candidate_list
observation
combat_result
inventory_changed
system_warning
```

### audience

同一事件对不同人显示不同文本。

```text
actor:
  你打开了木门。

target:
  张三打开了你面前的木门。

observer:
  张三打开了北面的木门。
```

因此事件应区分：

```text
actor_view
target_view
observer_view
admin_debug_view
```

### message_key

规则层只给 message_key，不写自然语言。

```text
message_key: transition.missing_item
message_key: transition.in_combat
message_key: death.enter_afterlife
message_key: command.ambiguous_target
```

### entities / values

所有可代入文本的对象和值都结构化传递。

```text
entities:
  actor: player_123
  item: ferry_ticket
  place: east_dock

values:
  fare: 50
  required_age: 16
  cooldown_left: 32s
```

Presentation 层负责决定显示名称、代词、数量、单位、颜色、强调方式。

### 内部错误与错误码

脚本错误、内容错误或内部 effect validation 失败时，不应把原始错误文本直接显示给普通玩家。系统应生成结构化错误事件，并给普通玩家一个可报告的短错误码。

```text
PresentationEvent:
  kind: system_error
  channel: system
  message_key: system.internal_error.report_code
  fields:
    error_code: MUD-8F3A2C
    contact_role: wizard
```

普通玩家看到：

```text
你感觉有什么东西出错了。错误码 MUD-8F3A2C。若你看到这个，请尽快联系巫师。
```

测试用户、巫师或开发者可以根据权限看到更多摘要：

```text
MUD-8F3A2C script error in scripts/items/old_lantern.lua:on_use
hook=on_use snapshot=content-42 reason=unknown effect move_entity
```

完整 stack trace、内部状态、脚本路径细节和 effect request 应进入服务器日志或管理工具，不直接暴露给普通玩家。

原则：

> 普通玩家看到可报告错误码；测试用户和巫师看到定位摘要；完整错误留在内部日志。错误显示策略由 actor/session 权限决定。

## 失败原因

失败原因应结构化，而不是硬编码文本。TransitionBlocker、RequirementFailure、ActionFailure 可以使用相同思想。

```text
Blocker:
  code
  source
  severity
  message_key
  fields
  hint_key?
  reveal_policy
```

例子：

```text
code: missing_item
message_key: transition.missing_item
fields:
  required_item: ferry_ticket
hint_key: transition.buy_ticket_at_dock

code: attribute_out_of_range
message_key: transition.age_too_low
fields:
  field: age_years
  required_min: 16
```

Presentation 负责渲染：

```text
你没有船票。
守卫只允许成年人进入。
```

## 隐藏条件与 reveal_policy

有些失败原因不应该直接暴露，否则会剧透谜题或隐藏机制。

```text
reveal_policy:
  always
  if_known
  admin_only
  never
```

例如隐藏镜门不应该说：

```text
你缺少 hidden_moon_key。
```

可以渲染成：

```text
镜面冰冷，没有回应。
```

同时 debug view 可以保留真实原因，方便 builder 检查。

## 多原因与刷屏控制

当多个失败原因同时存在时，普通命令默认只显示最重要的一个公开原因。

```text
普通命令:
  显示最高优先级 public blocker

inspect / routes / why:
  显示更多公开原因、候选目的地和提示

admin/debug:
  显示 debug blockers
```

原则：

> 系统可以保留完整结构化原因，但默认 Presentation 不刷屏。玩家主动检查时再展开。

## 候选列表

命令歧义、动态目的地、付费旅行、传送阵都可能返回候选列表。

候选应结构化：

```text
Candidate:
  id
  label_key
  entity_or_place
  available
  blockers?
  cost?
  risk?
  sort_key?
```

例子：

```text
ride cart:
  candidates:
    - place: luoyang.station
      available: true
      cost: 50
    - place: kaifeng.station
      available: false
      blocker: route_closed
    - place: yangzhou.station
      available: false
      blocker: not_enough_money
      required_money: 120
```

Presentation 可以渲染为简短列表，也可以在结构化客户端中显示为可选择按钮。

## 视角差异

同一底层事件应该根据接收者不同渲染不同文本。

例如 Transition 成功：

```text
event_type: transition_success
actor: player_a
from: east_dock
to: ferry_deck
method: board_ferry
```

渲染：

```text
actor:
  你登上了渡船。

source observers:
  张三登上了渡船。

destination observers:
  张三登上渡船，走上甲板。
```

规则层只产生事实，不决定每个视角的自然语言。

## 与其它系统的关系

### Command

命令解析、歧义、候选目标、失败原因都应产出结构化 payload。

```text
command.ambiguous_target
command.no_visible_target
command.requirement_failed
```

### Transition

Transition 的 destination、candidate、blocker、cost、warning 都是结构化字段。

```text
transition.blocked
transition.candidates
transition.confirm_required
transition.success
```

### Death

死亡结算应产出结构化结果。

```text
death.enter_afterlife
death.exp_lost
death.money_lost
death.item_lost
death.reincarnation_available
```

### Observation

房间描述、物品描述、装备观察、战斗观察也应结构化生成，再由 Presentation 渲染。

房间观察的具体结构在 `observation-system-notes.md` 中展开。核心是：房间保留独特基础描述，但时间、天气、光照、可见实体、出口、危险提示等动态上下文都应作为结构化 ObservationPayload 生成。

## 当前结论

```text
所有玩家可见文本都从结构化字段解析出来。
规则层不拼自然语言。
Action、Requirement、Transition、Death 只产出事实、原因、候选和事件。
Presentation/i18n 层负责文本、视角、语言、风格、隐藏条件和刷屏控制。
默认只显示最重要公开原因，主动 inspect / why / routes 才展开更多。
结构化 payload 也为未来客户端 UI 和国际化保留空间。
```

后续还需要讨论：

```text
message catalog 如何组织
实体显示名和代词如何处理
观察者视角如何选择文本模板
结构化客户端是否与文本客户端共用 payload
debug/admin presentation 如何避免泄露给普通玩家
```
