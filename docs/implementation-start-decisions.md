# 第一版实现启动参考

这份文档记录当前已经形成共识、足以指导第一版实现的参考边界。它不是设计冻结，也不表示所有高层问题都已经讨论完毕。目标是避免把已定共识反复翻出来细拆，同时继续只讨论真正影响长期代码骨架的问题。

未列入这里的玩法、数值、掉落、NPC、装备、UI 细节和脚本扩展，默认写到对应代码时再决定；如果某个问题会影响核心数据结构、持久化边界、world loop 或内容编译模型，再回到设计讨论。

## 目标

第一版目标是极小可玩 MUD，而不是完整游戏。

```text
验证：
  content load
  session / character binding
  world loop
  command execution
  entity relation
  observation / presentation
  SQLite persistence

不验证：
  战斗
  死亡
  副本
  掉落表
  装备词条
  完整 TUI
  大型脚本系统
```

每个可实现维度只做 2-3 个样例，用来形成对比、验证抽象、排除单例特例，不追求内容量。

## 最小可玩闭环

第一版实现：

```text
开发者临时身份登录
选择或创建角色
进入 2-3 个房间组成的小区域
look
go
say
get
drop
inventory
quit
断线重连
重启后恢复角色位置和背包
```

第一版内容：

```text
2-3 个房间
2-3 个出口/移动方向
2-3 个手工物品
2-3 种物品位置状态：房间地面、玩家背包、已丢弃地面物
```

## 持久化策略

第一版使用 SQLite。不要 event sourcing，不要完整世界快照，不要通用对象数据库。

SQLite 是第一版数据库后端选择，不是世界模型的一部分。持久化必须通过 `internal/persist` 的 Store 边界隔离，保留未来迁移到 PostgreSQL 或其它后端的能力。

```text
world / command:
  不直接依赖 SQLite
  不直接写 SQL

internal/persist:
  定义 Store interface
  实现 SQLite backend

future:
  可以替换为 PostgreSQL backend
  不应推翻 world / command / content 模型
```

只保存：

```text
accounts
characters
character_position
character_item_relations
item_instances
content_snapshot_version
```

第一版不保存：

```text
普通地面物
临时房间状态
副本
战斗状态
scheduled events
完整 live world
```

规则：

```text
玩家位置保存
玩家物品关系保存，第一版只使用 inside 表达背包/房间地面
玩家已确认物品实例保存
地上初始物品由 spawns.cue 重建
玩家 drop 到地上的物品第一版不做长期持久化
```

`reset_managed` 或定时刷新生成的低价值/工具性物品不追求资源守恒。玩家因为服务器重启、区域 reset 或定时刷新拿到多个同类实例是允许的，也可以作为长期规则存在。

```text
examples:
  树上的果子
  路边的野草
  前期山洞门口的灯笼
  临时帮助用火把
```

这些物品通常售价低、用途有限、主动获取容易，没有必要为了防止重复获取而持久化完整地面状态或追踪每次 reset 消耗。

不要把背包持久化为阻碍未来装备系统的简单数组。第一版虽然只实现 `inside`，但存储模型应保持 relation 思维：物品在背包里、地上、未来装备槽中，都是互斥位置关系。装备时物品不应仍然存在于背包里。

玩家看到“已经获得/已经移动/已经保存”的结果前，相关持久化必须已经成功；否则只能显示失败或 pending，不允许内存和数据库长期分叉。

第一版保存时机采用同步确认模型：改变玩家长期事实的命令，必须在展示成功前完成 `PersistStore` 写入。

```text
mutates player-interest state:
  persist before success presentation

observes or chats only:
  no persistence

persist fails:
  command fails
  do not show success
```

例子：

```text
get:
  validate pickup
  prepare item instance / relation change
  persist item instance + relation
  apply world mutation
  present success

go:
  validate transition
  persist character_position
  apply world mutation
  present new place observation

look / inventory / say:
  no persistent write
```

第一版不做异步保存、event sourcing 或 outbox。后续如果命令事务变复杂，再引入更完整的 durable command/event 流程。

第一版也不做通用 checkpoint 或后台周期性保存全部角色。可靠性来自同步确认保存，而不是定时补救。

```text
no first-version checkpoint:
  不周期性保存所有玩家状态
  不保存完整 live world snapshot
  不在 shutdown 时抢救大量未确认内存状态

graceful shutdown:
  停止接受新连接
  让 world loop 处理完已入队命令
  关闭 session
  flush/close SQLite resources
```

shutdown 是有序停止，不是补保存机制。未来如果引入长时间 buff、战斗中状态、副本 run、定时任务或 builder 草稿，再重新讨论 checkpoint。

## Go 模块边界

第一版 package 保持少而清楚：

```text
cmd/mudserver
  main.go

internal/content
  CUE Go API / Lua 加载
  ContentSnapshot
  EntityTemplate
  TagClass schema
  SpawnDefinition

internal/world
  World
  EntityID
  LiveEntity
  Relation
  PositionTransaction
  WorldMutation

internal/session
  Session
  Account/Character binding
  reconnect / kick old session

internal/command
  parse command
  execute look/go/say/get/drop/inventory/quit

internal/persist
  SQLite store
  load/save character position and inventory

internal/presentation
  structured events -> text output first
```

先不要拆更细。能写出来、能跑起来、能测试，比架构漂亮重要。

第一版依赖方向保持简单，重点避免循环依赖和边界穿透：

```text
session -> command -> world -> content
command -> presentation
cmd/mudserver wires persist, world, session, content, presentation together
```

边界约束：

```text
world:
  不依赖 session
  不保存 socket / connection 指针
  只关心 player entity、character id、recipient id 等稳定身份

session:
  不直接修改 world
  只提交 CommandRequest，接收 PresentationEvent

command:
  不直接操作 SQLite
  不直接拼中文
  不直接写 session socket
  不绕过 PositionTransaction 修改 relation

persist:
  保存长期事实
  不持有 world loop runtime object

presentation:
  渲染结构化 payload
  不修改 world mutation
```

第一版暂时不新增 `application` / `service` 总协调包。先由 `cmd/mudserver` 或很薄的启动/coordinator 代码串起 ContentCompiler、World、PersistStore、Session server 和 renderer。等代码开始膨胀，再提取应用层。

## World Loop 与命令执行

第一版使用单 world owner loop，顺序处理命令。

```text
Session input
  -> CommandRequest
  -> World loop
  -> Command handler
  -> WorldMutation / PresentationEvent
  -> flush to Session
```

第一版命令：

```text
look
go
say
get
drop
inventory
quit
```

第一版不做复杂 tick、战斗、死亡、副本、主动 AI、reaction cascade。后续 tick/scheduler 可以在 world loop 基础上加，不阻塞第一版。

第一版并发和写入边界：

```text
World loop:
  世界状态唯一写入口

Session / network goroutine:
  不直接修改 world
  只提交 CommandRequest

Command handler:
  在 world loop 内 resolve / validate / mutate / present

PositionTransaction:
  relation 改动唯一写入口

PresentationEvent:
  玩家输出唯一入口

PersistStore:
  玩家长期事实唯一写入口
```

第一版命令执行骨架：

```text
parse:
  raw input -> ParsedCommand

resolve:
  基于 actor 当前 place、可见 scope 和 relation 解析目标

validate:
  检查 actor / target / relation / tag requirements

mutate:
  通过 WorldMutation / PositionTransaction 修改状态

present:
  生成结构化 PresentationEvent
```

不要让 parser、session 层或 renderer 取得权威世界写权限。解析可以提前，权威 resolve、验证和执行必须在 world loop 中基于当前 live state 原子完成。

## Observation / Presentation 最小边界

第一版 `look`、`inventory` 等命令不拼接自然语言，只请求结构化观察结果。

```text
command/action:
  build ObservationContext
  request ObservationPayload

Observation:
  decides visible structured facts
  collects strong-tag ObservationFacet

Presentation:
  wraps payload as PresentationEvent

Renderer:
  concatenates text or renders TUI
```

第一版可以实现简单 text renderer，但它只是结构化 payload 的渲染器，不能反向约束规则层输出纯文本。

第一版只做：

```text
look current place
look visible target in same place
inventory
```

第一版不做：

```text
隐藏/反隐
鉴定/appraise/identify
光照限制
技能感知
距离衰减
完整 TUI layout
```

## 第一版 Transport

第一版实现使用 TCP line transport 作为调试和最小可玩入口。输入是普通命令行，输出由 text renderer 渲染结构化 `PresentationEvent` / `ObservationPayload`。

```text
TCP line session:
  read raw command line
  submit CommandRequest
  receive PresentationEvent
  render via text renderer
```

这只是第一版 transport 选择，不是核心协议边界。内部仍然使用结构化 command/event：

```text
Session:
  raw input -> CommandRequest
  PresentationEvent -> renderer/encoder

World / Command:
  不知道 TCP
  不知道 WebSocket
  不知道 TUI
```

长期仍保留 WebSocket + JSON / TUI 客户端方向。未来增加 `WebSocketJSONSession` 时，应复用 command、world、presentation，不改核心模型。

## 第一版验证策略

第一版测试要保护架构边界，不测试未来玩法。测试保持轻量，服务于恢复手写代码能力。

```text
Unit tests:
  ContentCompiler 校验
  World relation / PositionTransaction
  command parse / resolve 的纯逻辑

Integration tests:
  world + command + temp SQLite
  get/drop/go 后持久化事实正确
  restart/reload 后角色位置和玩家物品 relation 恢复

Manual smoke test:
  TCP line 连接
  look -> go -> get -> inventory -> drop -> quit
  restart server
  reconnect and verify persisted facts
```

优先保护的边界：

```text
ContentCompiler:
  entities key == id
  spawn references valid
  unknown tag rejected
  invalid script binding rejected

World relation:
  entity 同时只能在一个位置关系里
  get: room -> inventory
  drop: inventory -> room

Persistence:
  go 后重启，玩家位置恢复
  get 后重启，物品仍在玩家 relation 里
  reset-managed 默认物品可重建

Command:
  look / say 不写持久化
  get / drop 只通过 PositionTransaction 改 relation
```

第一版不做复杂 browser e2e、load test、完整 property testing、大型 golden text snapshot，或为未来系统创建大量 mock。

## ContentCompiler 最小 IR

实现顺序采用 `ContentSnapshot first, source loader second`。

```text
phase 1:
  hand-written ContentSnapshot fixture
  run world / relation / command / presentation / persistence

phase 2:
  CUE Go API loader produces the same ContentSnapshot type

phase 3:
  Lua script bindings attach to the same snapshot model
```

第一步可以先写：

```text
NewTutorialSnapshotForDev() -> ContentSnapshot
```

它硬编码 2-3 个房间、2-3 个物品和少量 spawns，用于先跑通核心世界闭环。World 不知道 snapshot 来自手写 fixture 还是 CUE。

第一版 ContentSnapshot 只需要：

```text
ContentSnapshot:
  version
  tag_classes
  entity_templates
  spawn_definitions
  script_bindings
```

EntityTemplate：

```text
id
display
tags
state_defaults
reactions
```

SpawnDefinition：

```text
id
entity
target
relation = inside
lifecycle = reset_managed
```

第一版内容源：

```text
CUE owns declarative content
Lua owns hot-loaded behavior
Go consumes normalized RuntimeContent
```

第一版不做：

```text
DropTable
复杂 resolver
复杂 policy registry
content migration
admin UI
```

ContentCompiler 是 source 到 runtime 的唯一入口：

```text
CUE / Lua source
  -> CUE Go API + Lua loader
  -> ContentCompiler semantic validation
  -> RuntimeContent IR
  -> ContentSnapshot
  -> World runtime
```

CUE 集成使用 `cuelang.org/go` 官方 API，不把 `cue` CLI 作为服务器运行时依赖。CLI 可以用于人工调试，但服务端内容加载不 shell out。

World loop、command handler、relation system 和 presentation system 只读取 `ContentSnapshot`，不直接读取 CUE AST、Lua 源文件路径或原始目录结构。source path、line/column、CUE package layout 只用于错误报告、hot reload validation 和开发诊断。

第一版加载校验只覆盖运行时骨架需要的边界：

```text
definition_id 唯一
entities key == entity.id
spawn.entity / spawn.target 引用存在
spawn.relation == inside
entity tags 已注册且符合 schema
reaction hook 已注册
script binding 可加载
new ContentSnapshot 与 live references 兼容
```

## 延后决定

以下内容写到对应代码或玩法时再决定：

```text
DropTable 形状
NPC 对话系统
装备词条生成
战斗公式
死亡惩罚数值
副本规则
更多 Lua primitive
复杂房间机关
完整 TUI 布局
builder/admin 工具
```

原则：

> 第一版先恢复手写代码能力。长期项目需要可演进边界，不需要在没有代码和反馈前把所有未来内容设计完。

## 仍可继续讨论的高层问题

以下主题尚未全部关闭，但讨论时只处理会影响代码骨架的部分，不深挖玩法细节：

```text
Observation / Presentation 的最小运行时边界
装备槽位与持有/穿戴 relation 的框架边界
最小持久化表与保存时机
World loop / command execution 的实现骨架
ContentCompiler IR 与加载校验的最小接口
Go package 边界是否足够支撑第一版
```
