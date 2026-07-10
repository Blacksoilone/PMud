# 设计讨论路线图

这份文档记录接下来需要逐个讨论的设计主题。它不是实现计划，也不是待办清单；它只是后续讨论的索引，防止我们在扩展想法时忘记框架真正需要先决定的边界。

## 当前实现参考

部分第一版共识已整理到 `implementation-start-decisions.md`，作为实现参考，不是设计冻结。剩余高层问题仍可继续讨论，但不再深挖可后置的玩法和 schema 细节。

继续原则：

```text
只在影响第一版代码骨架时继续讨论。
具体玩法、数值、掉落、NPC、装备、脚本扩展和 UI 细节，写到对应代码时再决定。
如果实现中发现文档不适合代码，以可维护代码为准，再回写文档。
```

第一版开工范围：

```text
极小可玩 MUD
2-3 个房间
2-3 个手工物品
look / go / say / get / drop / inventory / quit
Session / Character / PlayerEntity
ContentSnapshot
inside relation
SQLite 保存角色位置和背包
```

第一版明确后置：

```text
战斗
死亡
副本
DropTable
装备词条
完整 TUI
复杂脚本系统
```

仍需继续讨论的高层问题：

```text
Observation / Presentation 的最小运行时边界
装备槽位与持有/穿戴 relation 的框架边界
最小持久化表与保存时机
World loop / command execution 的实现骨架
ContentCompiler IR 与加载校验的最小接口
```

Go package 边界当前采用 `cmd/mudserver` 加 `internal/content`、`internal/world`、`internal/session`、`internal/command`、`internal/persist`、`internal/presentation`。第一版暂不新增 application/service 包。

第一版验证策略采用轻量三层：unit tests 保护 ContentCompiler/Relation/Command 纯逻辑，integration tests 跑 world + command + temp SQLite 核心闭环，manual TCP smoke test 走最小玩法。测试优先保护架构边界，不测试未来玩法。

当前已有文档：

- `mud-architecture-notes.md`：服务端运行时、内容定义、live state、持久化、热加载。
- `entity-effect-system-notes.md`：Entity、Tag、Provider、EffectiveView、Trigger、Effect、合成、贴附、随机能力的底层抽象。
- `loot-itemization-research.md`：Diablo、Wynncraft、ARPG 装备随机化调研。
- `mud-equipment-research.md`：MUD 装备、掉落表、文本可读性调研。

## 讨论原则

每个主题都先讨论框架边界，而不是具体内容设计。

应该优先回答：

```text
框架必须支撑什么？
哪些内容以后再设计？
哪些决定会把未来锁死？
哪些特例其实说明抽象缺了一层？
```

暂时不进入：

```text
完整实现计划
具体代码结构
具体地区、怪物、装备、技能数值
完整战斗系统
复杂合成/贴附内容
```

## 1. 命令系统

命令系统是 MUD 的入口。玩家输入自然语言式文本，底层应该转换为可验证、可排队、可持久化影响世界的动作。

需要讨论的问题：

```text
玩家输入如何解析？
look north / look door / open north 是否解析到同一个 entity？
go north / crawl into hole / enter portal 是否是同一种 Action？
命令是直接执行，还是先转成 Action？
Action 如何检查 Requirement？
失败反馈由谁生成？
命令解析如何处理别名、歧义、可见性、距离和容器？
```

期望产出：

```text
Command -> ParseResult -> Action -> Requirement -> Event/Effect 的基本模型
```

## 2. 空间与关系模型

我们已经倾向于把出口、容器、房间、物品、玩家、NPC 都作为 Entity，但还需要讨论 relation 的边界。

需要讨论的问题：

```text
located_in、contained_in、attached_to、equipped_by、worn_by、connected_to 是否统一建模？
房间和容器是否都是某种可容纳关系？
出口作为 entity 时如何连接两个 place？
尸体、背包、箱子、房间、玩家 inventory 有什么共同点和差异？
关系如何影响可见性、可触达性和命令目标解析？
```

期望产出：

```text
Entity relation 的最小集合，以及哪些关系必须持久化。
```

## 3. 时间系统

很多未来系统依赖时间：燃烧、腐烂、冷却、刷新、状态持续、NPC 行为、reset。

需要讨论的问题：

```text
server second 是否是固定 tick？
事件是每秒扫描，还是定时调度？
燃烧、腐烂、刷新、冷却如何推进？
离线时世界是否继续运行？
崩溃恢复后 timer 如何恢复？
世界暂停、区域休眠、无人区域如何处理？
```

期望产出：

```text
Tick / Timer / ScheduledEvent / WorldClock 的基本边界。
```

## 4. 持久化粒度

我们已经确定服务器重启不等于世界重置，但还需要决定第一阶段保存哪些状态。

需要讨论的问题：

```text
哪些 EntityInstance 必须保存？
哪些状态可以从 ContentDefinition 推导？
哪些 relation 必须保存？
掉落物、尸体、打开的门、烧了一半的火把是否保存？
玩家离线后角色是否仍在世界中？
世界状态如何 checkpoint？
是否需要事件日志，还是先只做 snapshot？
```

期望产出：

```text
ContentDefinition、LiveState、DerivedState、Snapshot/EventLog 的边界。
```

## 5. 内容文件格式与加载校验

内容数据必须能表达实体模板、出口/transition、spawn、基础 tag state 和 script binding，但第一版不应该过度复杂。DropTable 不进入第一版内容编译范围，等战斗、宝箱、采集或副本奖励真正需要时再设计。

当前结论：第一版内容源格式采用 CUE；Lua 只用于热加载行为脚本；Go ContentCompiler 读取 CUE、加载 Lua script bindings，并生成规范化 RuntimeContent / ContentSnapshot。数据库不是手写内容源，自制内容语言、YAML、JSONC 都不作为第一版主格式。

`definition_id` 采用 `kind.namespace.local_name`，例如 `room.tutorial.start`、`item.tutorial.old_lantern`。ID 只允许 lowercase ASCII、dot 分段、snake_case local name；玩家可见中文名称放在 display 字段，不混入 ID。

内容目录按 area 聚合：`content/schema/` 放全局 CUE schema，`content/shared/` 放跨区域复用内容和脚本，`content/areas/<area>/` 放该区域的 `area.cue`、`rooms.cue`、`items.cue`、`npcs.cue`、`spawns.cue` 和 area-local scripts。第一版 CUE 内容里完整手写 `definition_id`，不由 compiler 根据目录隐式补全。

EntityTemplate 使用强 tag 模型。Tag 不是布尔标签，而是类式语义对象：有 schema、属性、方法/provider 和事件响应。房间、物品、NPC 都是带不同 tags 的 entity；`component/core slot/extension` 只作为运行时编译布局术语，不作为内容作者面对的语义层。

Tag 方法通常由 Lua 组合 Go primitive：Lua 负责条件、选择器、概率和 effect request 组合，Go 负责核心执行、权限、事务和持久化。少数极底层、高频、完全核心的能力可以直接注册为 Go provider，不需要绕 Lua。

EntityTemplate 第一版统一使用 `entities` map。`rooms.cue`、`items.cue`、`npcs.cue` 只是文件拆分和作者便利，语义上都导出到 `entities`。map key 和内部 `id` 都完整手写，且必须一致。

Relation / spawn 第一版分三层：`EntityTemplate` 定义东西是什么，`SpawnDefinition` 定义初始化/reset 时生成什么并放到哪里，`LiveRelation` 记录具体实例现在在哪里。第一版只有最小 `inside` relation；每个 `SpawnDefinition` 只生成一个实例，不支持 `count` 或 spawn group。

需要讨论的问题：

```text
加载时要校验哪些引用？
热加载失败时如何保留旧 ContentSnapshot？
内容数据是否允许 include/template 复用？
```

期望产出：

```text
第一版内容数据的最小表达能力和加载校验规则。
```

## 6. 玩家身份、Session 与角色

网络连接、账号、角色、玩家实体不应该混在一起。

当前结论：第一版采用混合模式。架构上保留 `Session / Account / Character / PlayerEntity` 的完整边界；开发期允许 anonymous/dev session 自动创建临时账号和临时角色。

断线采用短暂保留模型。Session 断开后，PlayerEntity 短暂留在世界中并进入 reconnect grace window；窗口内重连可恢复控制。窗口超时后按离线可利用性结算：安全场景 logout；如果离线会让玩家显著获利或规避代价，战斗、副本、特殊环境或特殊状态可显式声明 fallback、失败或死亡。高危区域本身不构成断线死亡条件。

重连绑定规则：同一个 `Character` 同一时间只能有一个 controlling `Session`。同账号新连接可以踢掉旧连接并接管；不同账号不能绑定同一角色，除非是显式 admin/debug/observer 权限。

账号角色规则：一个 `Account` 允许拥有多个 `Character`。MUD / MMO 经常有永久性或半永久选择，玩家不应为了体验不同玩法而创建多个账号、邮箱或认证身份。

dev / anonymous 身份规则：临时账号和临时角色只用于开发者测试，不是普通玩家游客模式；不能升级、迁移或绑定到正式账号，将来引入邮箱或外部认证后也不迁移。

需要讨论的问题：

```text
身份/session 主题已形成第一版结论。
```

期望产出：

```text
Session、Account、Character/PlayerEntity 的边界。
```

## 7. 观察、可见性与信息披露

我们已经多次提到 tag 可以根据观察者能力显示不同语义。这需要成为统一系统，而不是散落在命令里。

需要讨论的问题：

```text
look、examine、appraise、identify 有什么区别？
玩家默认看到哪些信息？
技能、工具、距离、光照、状态如何影响可见性？
tag 的展示语义由谁生成？
隐藏出口、伪装物品、未知属性如何表达？
```

期望产出：

```text
ObservationContext -> Presentation 的基本模型。
```

## 8. 最小装备与掉落框架

装备内容本身以后设计，但框架需要支撑来源掉落表、装备槽位、实例持久化和简洁展示。

需要讨论的问题：

```text
第一版装备是否只支持武器/防具/手持物？
装备槽位是 component 还是 relation？
DropSource / DropTable 第一版需要支持哪些来源？
Loadout 和 DeathDrop 是否分开？
物品实例是否从一开始保存 provenance 位置？
compare 是否第一版就需要，还是只留接口？
```

期望产出：

```text
装备框架最小能力，不决定具体装备内容。
```

## 9. 第一版最小闭环

在上述边界讨论完成后，再收束第一版目标。第一版应该小到能手写、能理解、能验证。

当前结论：第一版选择“极小可玩 MUD”，但必须相当小。每个能实现的维度都只做 2-3 个样例，用来形成对比、验证抽象并排除单例特例，而不是追求内容量。

范围控制原则：

```text
不是 1 个：
  单例容易把特殊情况误认为通用规则

不是很多个：
  内容量会掩盖框架问题，拖慢实现

通常 2-3 个：
  足够产生对比
  足够验证 relation / command / observation / persistence
  足够发现“只支持一个例子”的硬编码
```

候选最小闭环：

```text
TCP 多人连接
Session read/write pump
World loop
Entity + relation 雏形
2-3 个房间
2-3 个出口/移动方向
look / say / go / get / drop / inventory / quit
2-3 个手工物品
2-3 种物品位置状态：房间地面、玩家背包、已丢弃地面物
持久化玩家位置和少量 live state
```

第一版明确不做：

```text
正式战斗
正式死亡
副本
装备词条生成
复杂掉落表
账号邮箱验证
完整 TUI 图形布局
脚本系统
大型内容区域
```

期望产出：

```text
第一版实现范围，以及明确不做什么。
```

## 推荐讨论顺序

建议按以下顺序逐个推进：

```text
1. 命令系统
2. 空间与关系模型
3. 时间系统
4. 持久化粒度
5. 内容文件格式与加载校验
6. 玩家身份、Session 与角色
7. 观察、可见性与信息披露
8. 最小装备与掉落框架
9. 第一版最小闭环
```

原因是：命令系统决定玩家输入如何进入世界；关系模型决定所有对象如何相互连接；时间和持久化决定世界是否真实存在；内容格式决定能否被作者维护；其余系统在这些基础上展开。

## 下一步

下一轮建议从 **命令系统** 开始。

第一个问题可以是：

```text
look north、open north、go north、crawl into hole 这些输入，底层应该如何解析到 Action 和 Entity？
```
