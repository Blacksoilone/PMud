# 实体、Tag 与效果系统设计笔记

这份笔记记录前期讨论中形成的底层抽象原则。它不是实现计划，也不是第一版功能清单。它的目的，是确保第一版服务端虽然很小，但底层不会阻断未来的组合式玩法、持久世界、合成、贴附、随机道具和特殊神器。

核心目标是：**底层越干净越好，特例越少越好。特殊内容可以存在，但特殊内容必须通过通用机制表达。**

## 最高原则

### 模拟直觉，不模拟物理

这个项目不是物理引擎，也不是全量世界模拟器。底层抽象的目标不是让每一滴液体、每一粒粉末、每一层容器、每一次热传导都被真实模拟，而是让玩家感觉世界符合直觉、可推理、可交互。

因此，系统应该优先表达游戏语义：

```text
玩家觉得能做的事，大体应该能做。
玩家看到的性质，应该能被信任。
复杂互动应该通过清晰规则出现，而不是通过无限递归模拟涌现。
```

例如：

```text
血瓶可以放进袋子，因为玩家拿到的是瓶子。
血液仍然是液体，但它作为瓶子的内容状态存在。
鱼油贴到衣服上可以提供腥味、滑溜溜、易燃等表现，但不会让衣服变成液体。
袋子不能装袋子，因为递归容器会造成复杂度和服务器风险。
```

原则：

> 框架要模拟一个符合直觉的文字游戏世界，而不是实现真实物理。优先选择玩家能理解、系统能控制、内容作者能维护的规则。

### 特殊内容不能变成架构特例

一件全游戏唯一的神器可以拥有非常特殊的效果，但它不应该在核心系统里开专门分支。

坏设计：

```text
if weapon.id == "space_twist_staff" {
  killWithoutDrop(enemy)
}
```

好设计：

```text
时空扭曲法杖:
  tags:
    - 法杖
    - 时空扭曲

时空扭曲:
  trigger: on_attack_hit
  chance: 5%
  effect:
    - banish target to twisted_space
    - suppress corpse/drop generation
```

即使 `时空扭曲` 全游戏只出现在这一把武器上，它也应该是一个正常 tag、trigger、effect 组合，而不是核心战斗逻辑中的特殊 if。

原则：

> 内容可以独一无二，但实现路径不能独一无二。独一无二的物品应该由通用的 tag、trigger、requirement、effect、modifier、relation 和 event 组合出来。

### 如果需要特例，要怀疑抽象

如果某个内容需要：

```text
特殊字段
特殊保存逻辑
特殊命令路径
特殊战斗分支
特殊房间类型
特殊物品类型
```

就要先问：

```text
这是内容真的不能归入体系？
还是底层抽象缺了一层？
```

大多数时候，应该新增通用机制，而不是新增内容专用分支。

例如，为时空扭曲法杖增加 `on_attack_hit` trigger 是通用机制扩展；在攻击代码里检查这把法杖的 ID 是架构污染。

## Entity 是底层统一实体

旧设计里曾经分为：

```text
玩家
NPC
房间
物品
容器
```

这个分类会令人别扭，因为它混合了不同维度：

```text
玩家 / NPC    是控制来源或行动者类型
房间          是空间节点
物品          是可操作实体的常见表现
容器          是一种能力
```

更干净的底层应该是：

```text
Entity
  + tags
  + relations
  + providers
  + state
```

旧分类只是 archetype 或查询结果：

```text
玩家 = Entity + Actor + PlayerControlled + Located + Inventory + Perception
NPC  = Entity + Actor + AIControlled + Located + Dialogue?
房间 = Entity + Place + Container-like spatial contents + Environment
物品 = Entity + Physical + Portable?
容器 = Entity + Container
出口 = Entity + Exit capability
```

原则：

> 不用根类型回答“它是什么”，而用能力和关系回答“它能参与哪些规则”。

这允许房间、门、出口、火焰、状态源、贴附物、容器、NPC 和玩家都进入同一套持久化、观察、命令解析和效果系统。

## 出口也是实体

方位不应该是特殊模型。`north`、`狗洞`、`裂缝`、`传送门`、`楼梯` 的区别主要是显示、别名和交互文本不同，而不是底层类型不同。

出口应该是房间中的一个实体，并具备 `Exit` 能力：

```text
exit_001:
  tags:
    - exit
    - north
    - wooden_door
  display:
    name: 北边的木门
    aliases: [north, n, door, wooden door]
    travel_text: 你向北走去。
  exit:
    from: room_a
    to: room_b
```

狗洞也是出口：

```text
exit_002:
  tags:
    - exit
    - hole
    - narrow
  display:
    name: 墙角的狗洞
    aliases: [hole, dog hole]
    travel_text: 你趴下来钻进狗洞。
  exit:
    from: room_a
    to: room_c
```

这样玩家可以自然地：

```text
go north
open door
unlock north
crawl into hole
look hole
burn door
throw torch through hole
```

底层都是在操作一个具备 `Exit` 能力的 entity。

## 锁和通过条件属于动作约束

锁不是出口的专用特例，通过条件也不是出口的专用特例。它们都属于更大的抽象：**动作约束**。

一个出口可以有很多动作：

```text
traverse
open
unlock
force
crawl
climb
inspect
```

每个动作可以有若干 requirement：

```text
traverse:
  - requires unlocked
  - requires body_size <= small
  - requires posture = crawling
  - requires climb_skill >= 30
  - requires not overloaded

unlock:
  - requires has_item brass_key

force:
  - requires strength_check >= 20
```

所以概念上应拆成：

```text
Exit
  连接 from -> to

State
  locked/open/blocked/hidden/damaged/collapsed

Action
  traverse/open/unlock/force/crawl/climb/inspect

Requirement
  执行动作前必须满足的条件

Effect
  动作成功或失败后的结果
```

传统方位出口只是最简单的出口 archetype：

```text
plain_direction_exit:
  tags: [exit, direction, north]
  requirements: []
  default_action: traverse
```

## Tag 是玩家可学习、可信任的语义单位

tag 不是随便的字符串，也不是所有数值组合的名字。tag 对玩家来说是一种承诺。

Tag 也不是和面向对象冲突的“裸数据”，更不是简单布尔参数。更合适的理解是：Tag 是一种可组合的类式语义对象。

```text
Entity:
  对象身份

TagClass:
  定义一种能力、性质、接口和行为约束
  有 schema、属性、方法、provider 和事件响应

TagInstance / TagState:
  某个具体 Entity 上的 tag 实例状态
  保存该能力在这个对象上的当前属性值

Behavior / Provider:
  TagClass 暴露给引擎和其它系统的受控方法

Service:
  跨对象事务如何协调
```

也就是说，tag 是对象的可组合能力模块；Behavior/Provider 是 tag 的面向对象接口；Service 负责跨对象事务。

内容层不需要把 `Component` 作为和 `Tag` 平级的概念。内容作者面对的是强 tag：

```text
place:
  kind: room
  capacity: players_allowed
  methods/providers:
    Observation.Facets
    Transition.Exits

portable:
  methods/providers:
    Inventory.CanPickUp
    Inventory.Weight

light_source:
  radius: 2
  methods/providers:
    Observation.LightContribution
    Effect.SetLit
```

`Component`、`core slot`、`extension map` 是运行时布局和优化术语，不是内容语义边界。

错误方向是把 tag state 当成外部系统可以任意修改的裸字段：

```text
entity.tags["health"].current -= 10
entity.tags["position"].place = new_place
entity.tags["inventory"].items.append(item)
```

推荐方向是让重要能力暴露受控行为：

```text
actor.Health().ApplyDamage(damage)
actor.Position().CanMove(context)
actor.Inventory().CanAccept(item)
item.Durability().Damage(amount, reason)
place.Observation().Facets(observer, mode)
```

跨对象操作由 service 处理：

```text
TransitionService.Execute(plan)
InventoryService.PickUp(actor, item, place)
EquipmentService.Equip(actor, item)
DeathService.Resolve(actor, context)
```

这样可以同时保留组合能力和对象不变量。

例子：灯笼。

```text
Entity: lantern_01
  Tags:
  portable
  light_source
  fuel_consumer
  fragile
  flammable
  container_like_oil_holder

Behaviors:
  LightSource.Illuminate(context)
  FuelConsumer.Consume(duration)
  Fragile.ApplyImpact(force)
  OilHolder.Refill(oil)
```

如果用传统继承表达，很容易变成：

```text
MagicalFuelConsumingEquippableFragileQuestLantern
```

强 tag 负责避免继承爆炸；面向对象行为负责避免状态被随意修改。

### Tag 的运行时性能原则

Tag 适合表达内容和组合能力，但不应该成为热点路径上的查询算法。

不推荐在运行时频繁这样做：

```text
每次攻击时遍历武器所有 tag
每次观察时遍历实体所有 tag
每次移动时遍历房间所有 tag
每个 tick 扫描全世界对象寻找某个 tag
```

推荐模型是：

```text
内容层:
  使用强 tag 表达能力、性质、触发器和状态。

加载或状态变化时:
  把 tag 编译成 runtime view。

热点路径:
  使用 flags、typed tag accessor、trigger index、provider list。
```

也就是说：

```text
Tag 是内容表达。
Runtime index 是执行结构。
热点路径不遍历全部 tag。
```

第一版可以按这个方向设计：

```text
Entity:
  id
  flags bitset for common capabilities
  typed slots for common tag states
  extension map for rare tag states
  trigger index by trigger kind
  provider lists for observation / transition / effect
```

常见能力应是 O(1) 检查或强类型访问：

```text
entity.Flags.Has(Flammable)
actor.Health().ApplyDamage(...)
actor.Inventory().CanAccept(item)
place.Observation().Facets(observer, mode)
weapon.Triggers(OnAttackHit)
```

罕见能力可以通过扩展 tag state 或 provider 查找，但不应污染高频路径。

### 性能分层不能泄露语义

core slot、extension tag state、flag、index、provider list 都是运行时效率策略，不是游戏语义边界。

同一个能力无论存放在 core slot 还是 extension map，都应该能正常使用。移动到 core 只是为了更快、更直接、更容易维护热点路径。

原则：

```text
语义稳定:
  外部系统通过 typed accessor / behavior 使用能力。
  不依赖能力当前存放在 core slot 还是 extension tag state。

存储可迁移:
  extension 将来变常用，可以晋升为 core slot。
  core slot 如果后来证明不常用，也可以退回 extension。

调用方不感知:
  调用 actor.Durability() / item.LightSource() / place.Fallback() 不应因为存储位置变化而改代码。
```

例子：`durability` tag 早期可以是 extension。

```text
content:
  tags:
    durability:
      max: 100

runtime early:
  extension["durability"]

access:
  item.Durability()
```

如果后期耐久参与大量战斗、死亡、修理、装备损坏规则，可以晋升为 core slot。

```text
runtime later:
  entity.durability core slot

access still:
  item.Durability()
```

迁移时 loader 可以把同一份内容定义编译到新的 runtime slot；外部语义和内容写法不需要大改。

硬原则：

> 为了效率引入的人为分层，不能成为内容语义分层。优化边界必须被 accessor、behavior 和 loader 隔离起来。

复杂度原则：

```text
普通命令:
  O(当前房间相关对象 + 当前命令相关组件)

观察:
  O(当前房间可见对象 + 当前房间 observation providers)

攻击:
  O(攻击者 / 武器 / 目标 / 房间相关 trigger list)

移动:
  O(当前出口 / transition + 当前房间 enter reactions)

tick:
  O(active set)，不是 O(world)
```

硬原则：

> 任何热点路径上的 O(N) 都必须说明 N 是什么，并且 N 应该是局部集合，例如当前房间对象数、当前 trigger list 长度、当前 active combat 数，而不是全世界实体数。

这些具体实现手法，例如 bitset、typed slot、trigger index、provider list，可以等真正写到这一步时再详细展开。

如果玩家看到 `可燃`，他可以不知道具体数值，但必须能相信：

```text
可燃 = 能被点燃，并按燃料/燃烧强度消耗
```

它可以烧得快、烧得慢、烧得旺、烧得久，但不能点燃后喷水。如果点燃后喷水，就应该是另一个 tag，例如：

```text
遇火喷水
伪可燃
火焰反应
水囊
```

原则：

> 属性差异可以留在同一 tag 内；核心行为差异必须拆成不同 tag。系统允许代码复用，但不允许语义欺骗。

同名 tag 的核心行为必须一致，属性可以不同。

## 底层 Tag 与玩家可见语义分离

tag 不一定直接暴露给玩家。底层 tag 更像机制单位；玩家看到的是这个 tag 根据内部数值、观察者能力和上下文投影出来的表现。

例如底层可以只有一个 `可燃` / `Combustible`：

```text
Combustible:
  fuel_amount
  burn_intensity
  ignition_point
  flame_temperature
  light_intensity
```

玩家不一定直接看到：

```text
tags: 可燃
```

而是看到展示层生成的语义：

```text
flame_temperature < 0:
  看起来燃着一团冷焰。

0 <= flame_temperature < 100:
  这东西能发出温热的火光，但似乎点不燃普通燃料。

flame_temperature >= 100:
  这东西能正常燃烧，应该可以点燃干燥材料。
```

这样系统不需要为了每个数值阈值拆底层 tag。`冷焰`、`暖火`、`烈焰`、`好像很难熄灭` 可以是展示语义或 perceived facet，而不是独立机制 tag。

需要拆底层 tag 的情况，是事件响应机制真的不同：

```text
普通燃烧:
  消耗 fuel_amount，产生热和光

遇火喷水:
  被点燃时产生水流，不进入燃烧循环

爆燃:
  点燃瞬间爆炸，之后不持续燃烧

永燃:
  不消耗 fuel_amount，或消耗规则完全不同
```

原则：

> 机制相同、数值导致表现不同，优先通过展示层表达；机制不同、事件响应不同，才拆底层 tag 或 behavior provider。

这也修正了“tag 是玩家可学习语义单位”的说法：玩家学习的是系统呈现出的稳定语义，而底层可以用更少、更干净的机制 tag 支撑这些表现。

## TagDefinition 与 TagState

tag 不应该只是字符串。它更像一种能力定义或性质定义。

应区分：

```text
TagDefinition
  定义这个 tag 是什么
  定义它响应哪些事件
  定义它有哪些属性 schema
  定义它暴露哪些方法/API
  定义它如何展示给玩家
  定义它如何传播、冲突、衰减、结算

TagState
  某个具体实体上，这个 tag 的当前属性值
```

例如 `可燃` 定义只有一份：

```text
可燃 TagDefinition:
  属性 schema:
    fuel_amount
    burn_intensity
    ignition_point
    flame_temperature
    light_intensity
  行为:
    on_ignite
    on_server_second
    on_extinguish
  展示:
    根据 fuel_amount 显示模糊描述或精确数值
```

但每个实体上的 `可燃` 状态不同：

```text
干树叶:
  可燃:
    fuel_amount: 10
    burn_intensity: 1

火龙心脏:
  可燃:
    fuel_amount: 1000000
    burn_intensity: 80
```

持久化保存的是 `TagState`，不是行为代码。

### Tag 方法

Tag 的“方法”应该理解为受控行为 API，而不是裸字段函数。推荐执行链是：

```text
Lua method / reaction:
  组合条件、选择器、概率和 effect primitive
  调用 Go 暴露的核心能力
  返回 effect requests

Go primitive / engine service:
  实现真正的核心动作
  验证权限、状态、位置、事务和持久化边界
  通过 WorldMutation 或专用 service 提交变化

Pure Go provider:
  少数极底层、高频、完全核心的能力
  不需要绕一遍 Lua
```

也就是说，Tag 方法通常可以由 Lua 组合，但 Lua 调用的是 Go 写好的核心内容。

例子：`portable.pickup` 可以是 Lua 组合入口。

```text
portable.pickup(ctx):
  check conditions via Go helpers
  return effects.move_item(ctx.self, ctx.room.ground, ctx.actor.inventory)
```

真正移动物品仍由 Go 的 `PositionTransaction` 执行。

例子：`place.available_transitions` 更适合直接 Go provider。

```text
place.available_transitions(entity):
  core Go provider
  used by go/look/path systems
  no Lua dispatch on hot path
```

判断标准：

```text
适合 Lua method/reaction:
  内容差异大
  需要热迭代
  主要是组合已有 primitive
  例如特殊物品使用、房间机关、NPC 对话分支

适合 pure Go provider:
  极底层核心语义
  高频路径
  事务/权限/持久化边界很强
  完全没必要给内容作者热改
  例如唯一位置约束、基础容纳关系、核心 transition 枚举
```

原则：

> Tag methods are behavior APIs. Lua may compose method behavior by calling Go primitives; Go owns core execution. Very low-level core semantics can be pure Go providers without Lua dispatch.

## 可见性规则保持朴素

tag 的内部属性可以有显示门槛，但不需要一开始设计复杂的信息层级。

例如 `可燃` 可以先用三档模糊描述：

```text
fuel_amount 很低:
  很快就烧完了

fuel_amount 中等:
  能烧一会儿

fuel_amount 很高:
  好像很难熄灭
```

如果观察者满足详细查看条件，则直接显示数据：

```text
fuel_amount=1500
burn_intensity=123
```

原则：

> 默认给玩家自然语言判断；达到条件后给准确数值。模糊描述可以不精确，但不能误导。

## 属性不能全塞进 Tag，也不能全塞进 Entity

前期讨论中出现过两个极端：

```text
所有属性都属于 tag
所有属性都属于实体
```

两个方向都有问题。

如果所有属性都属于 tag，`剑` 下面有伤害、锋利度、长度、重量；这些属性一旦出现负数或极端值，就会诱导拆出大量 tag，例如：

```text
轻长钝回复剑
```

这会造成组合爆炸，丧失组合系统的价值。

如果所有属性都属于实体，又会要求每个实体预先拥有大量无关属性。一个铁棍本身没有燃烧时间、火焰温度、光强、烟雾量；只有贴附涂油布条后，它才表现出可燃能力。让所有物体都预定义所有潜在属性会导致数据极度膨胀。

更合适的模型是：

> 只有具备某种能力的实体、tag state、组件、状态或关系，才携带该能力所需的属性。其他对象没有这些属性，也不需要默认值。

## Capability Provider 与 EffectiveView

一个实体的最终表现，不一定只来自它自身。它可以来自：

```text
自身组件
自身 tag state
装备
贴附物
容纳物
状态效果
环境影响
关系投影
```

因此需要一个概念：`Provider`。

```text
Provider
  来自 entity 自身、tag state、attachment、equipment、status effect、environment
  提供 tags、capabilities、stats、behaviors、modifiers
```

系统操作时不直接看原始实体，而是看解析后的 `EffectiveView(entity)`：

```text
EffectiveView
  汇总所有 provider
  计算 effective tags
  计算 effective capabilities
  计算 effective stats
  解决冲突
  生成命令、观察、战斗、合成、贴附使用的最终视图
```

例如铁棍贴附涂油布条：

```text
铁棍:
  tags: [棍状, 金属, 固体]
  physical:
    length: 80
    weight: 3

涂油布条:
  tags: [可燃, 粘附性]
  combustible:
    fuel_amount: 500
    burn_intensity: 20
    flame_temperature: 180
    light_intensity: 30

关系:
  涂油布条 attached_to 铁棍
```

铁棍本身不需要燃烧字段。但通过 attachment provider，`EffectiveView(铁棍)` 可以表现为：

```text
tags: [棍状, 金属, 固体, 可燃]
capability: Combustible
fuel_amount: 来自涂油布条
```

燃料烧完时，消耗的是涂油布条上的运行状态，而不是铁棍本体。

## 合成与贴附

合成和贴附都应该建立在 tag、capability、provider 和 relation 上，但二者语义不同。

### 合成

合成是材料匹配和产物生成。

材料不是具体物品，而是 tag/capability 条件：

```text
火把配方:
  材料:
    handle: requires [棍状]
    fuel: requires [可燃, 粘附性]
```

任何对象只要其 effective tags 是材料要求的超集，就可以作为该材料使用。

但配方不必只检查 tag 是否存在。更深的配方可以检查 tag 或 capability 内部的属性 predicate。

例如普通火把可以只检查：

```text
handle:
  has: 棍状

fuel:
  has: 可燃
  has: 粘附性
  guard:
    fuel_amount > 0
```

而高级配方可以检查更具体的条件：

```text
寒灯配方:
  fuel:
    has: Combustible
    requires:
      flame_temperature between -30 and -5
      fuel_amount >= 800
      burn_intensity between 5 and 15

  vessel:
    has: 稳定载体
    requires:
      material_stability >= 60
```

这些条件应该检查 `EffectiveView`，而不是只检查实体原始数据。这样，一个铁棍本身没有可燃能力，但贴附涂油布条后，`EffectiveView(铁棍)` 可以提供 `Combustible`，从而满足普通火把配方。

配方也可以限制 provider 来源：

```text
普通火把:
  accepts provider source: intrinsic OR attached

炼金核心:
  requires provider source: intrinsic

献祭仪式:
  requires provider source: living_organ
```

玩家可见的展示语义可以作为内容作者的简写，但规则最终应落到底层 predicate。例如：

```text
requires facet: 冷焰
```

可以被系统展开为：

```text
has Combustible
and Combustible.flame_temperature < 0
```

这样玩家可以通过“冷焰”理解世界，内容系统则保持稳定的机制条件。

合成产物可以从材料获得属性，但不应该理解成 OO 继承。更准确地说，是配方声明材料能力如何转移、转换或保留来源关系。

三种模型：

```text
复制模型:
  材料被消耗，产物复制部分 tag state

转换模型:
  配方按公式从材料属性生成产物属性

组成模型:
  材料成为产物内部组成部分，来源关系继续存在
```

第一版可以优先考虑转换模型，长期保留组成模型空间。

### 配方条件层级

配方复杂度不应该均匀增加。高级配方不意味着所有属性都变成严格窄区间。更好的原则是：**普通配方防荒谬，高级配方考理解，神器配方考推理和流程。**

推荐分层：

```text
日常配方:
  主要检查 tags
  只做少量边界 guard，防止荒谬结果

进阶配方:
  tags + 1-2 个关键属性
  大多数条件仍然宽松

专门配方:
  少数严格属性 + 多数宽松边界
  需要玩家理解材料性质

神器配方:
  极端苛刻条件
  复杂属性配平
  限定材料或限定来源
  需要游戏内线索和高技能观察能力
```

例如普通火把只需要：

```text
棍状物 + 可燃粘附物
```

它最多检查燃料量不是 0、火焰表现不是完全相反的极端状态。玩家不应该为了日常物品精确配平燃烧强度和光强。

而神器配方可以要求：

```text
限定材料:
  灰塔守门人的钥斧碎片
  月蚀夜采集的冷焰
  未被点燃过的龙心余烬

复杂条件:
  flame_temperature 低于 0，但不能冻结容器
  fuel_amount 超过某阈值
  burn_intensity 被压到很低
  darkness.intensity 足够高
  材料来源没有被污染
  合成地点处于特定环境
```

神器要求可以苛刻，但不应该靠瞎试。游戏内必须提供可推理线索：

```text
某个 NPC 知道一部分
某本书记载一部分
失败反馈暗示一部分
旧遗物描述暗示材料来源
区域环境暗示温度、时间或仪式条件
高技能玩家能直接看见相关数值
```

原则：

> 配方隐藏的不是随机答案，而是可推理的结构。

复杂配方是内容设计，不是日常负担。底层支持属性 predicate，但普通玩法不应滥用它。

### 贴附

贴附是关系，不是一次性变换。

```text
红色染料 attached_to 布衣
鲑鱼鱼油 attached_to 布衣
```

宿主保留自身语义，附着物通过 provider 改变 `EffectiveView`。这件事应优先通过普通 tag 的添加、屏蔽、投影和冲突解决来表达，而不是默认引入复杂的来源、品质、稳定性或衰减字段。

```text
布衣自身:
  固体
  可穿戴

红色染料传播:
  红色

鲑鱼鱼油传播:
  腥味
  滑溜溜
  高度易燃

不传播:
  粉状
  液体
```

贴附的价值不是“免费强化”，而是让玩家消耗一个物品，给另一个对象补足某些性质，从而满足配方、仪式、通行、使用或后续操作要求。

例如仪式匕首需要：

```text
材料 A:
  has 匕首
  has 银质

材料 B:
  has 血液
  has 诅咒 OR 赐福
```

玩家如果只有普通鸡血，可以选择狩猎强大的诅咒/神圣生物取得正统血液，也可以把伪典残页贴附到血瓶上：

```text
鸡血:
  血液

伪典残页:
  诅咒
  人造物
  邪典

贴附后 EffectiveView(鸡血瓶):
  血液
  诅咒
  人造物
  邪典
```

普通配方可能只检查 `血液 + 诅咒`，因此可以使用。更严格的配方可以要求：

```text
has 血液
has 古老诅咒
not has 人造物
```

这样贴附提供了替代路线，但不会破坏正统稀有材料的价值。区别通过 tag 组合表达，而不是为每个 tag 都设计复杂 provenance。

原则：

> 贴附优先通过添加、屏蔽或投影 tag 来改变 EffectiveView。不要默认引入复杂来源、稳定性、品质字段。只有当某类玩法反复需要更细粒度区分时，才升级为结构化属性。

贴附关系必须持久化，否则重启后涌现内容会丢失。

## 冲突组和传播规则

有些 tag 可以共存，有些 tag 属于冲突组。

例如物态：

```text
group: phase
members:
  固体: priority 100
  液体: priority 80
  粉状: priority 70
  气体: priority 60
rule:
  keep highest priority
  tie: prefer host
```

颜色可能是冲突组，也可能允许混合；气味通常可以叠加；易燃性可能更适合用数值聚合。

原则：

> 冲突规则属于 tag 体系，不应该散落在每个命令或每个物品里。

## Action、Trigger、Requirement、Effect、Event

底层应该能表达以撒式的扩展性：手工设计的独特物品和随机生成的怪东西，都落到同一套机制里。

核心概念：

```text
Action
  玩家或系统尝试做的事
  例如 attack、use、traverse、inspect、ignite、attach

Requirement
  动作执行前必须满足的条件

Trigger
  某个时机
  例如 on_attack_hit、on_use、on_server_second、on_enter_room

Effect
  触发后产生的效果
  例如 deal_damage、heal、move_entity、banish_entity、create_entity

Event
  世界已经发生的事实
  例如 AttackHit、PlayerMoved、ItemDropped、DoorUnlocked

Modifier
  修改 EffectiveView 或结算结果
```

攻击不应该直接关心神器。攻击产生事件，tag/effect 响应事件：

```text
Attack(actor, target, weapon)
  -> AttackHit

锋利       -> 修改伤害分配
火焰       -> 添加燃烧
吸血       -> 把部分伤害转成治疗
时空扭曲   -> 概率放逐目标
诅咒       -> 可能反噬使用者
```

## 学习以撒的结合：效果比物品更底层

《以撒的结合》的重要启发不是某个具体道具，而是底层允许道具效果作为可组合行为修改器存在。

手工道具可以极其独特：

```text
R 键
创世记
达摩克利斯之剑
```

随机或异常道具也可以由底层效果拼装出来。关键是：

```text
手工设计物
随机生成物
合成产物
贴附变异物
```

在运行时都应该落到同一套：

```text
Entity + Strong Tags + Providers + Triggers + Effects + Relations
```

这样系统可以同时支持：

```text
创作者亲手设计一个全游戏唯一的神器
系统生成一个从未手写定义过的奇怪物品
```

## 随机道具生成：模板、积分、独占组和审计

组合式底层会让随机道具变得非常有趣。随机不再只是：

```text
伤害不同的铁剑
```

而可能是：

```text
燃烧的铁剑
能把敌人变成青蛙的飞镖
挥动时回复耐力的负重量剑
```

但这也很容易生成摧毁游戏的神器。尤其 MUD 是持久世界，问题比单局 roguelike 更严重。

因此随机生成不能从全体 tag/effect 中无约束抽取。应该至少有五层约束。

### 1. 模板白名单

先选设计者预定义的物品模板，再从模板允许的 tag/effect/modifier 池里随机。

```text
铁剑模板:
  base tags: [剑, 金属, 可挥舞]
  allowed_random_tags:
    - 锋利
    - 沉重
    - 燃烧
    - 冰冷
    - 吸血
    - 易碎
    - 诅咒
  forbidden_random_tags:
    - 复制自身
    - 全局传送
    - 重置世界
```

模板不是实例，而是生成边界。它保证随机结果不越过物品族群的语义。

### 2. 积分预算

每个 tag/effect/modifier 有成本。不同掉落来源有不同预算：

```text
普通野外掉落: 低预算
普通副本产出: 中预算
Boss 特殊产出: 高预算
手工神器: 可突破预算，但必须显式设计
异常生成: 可越界，但需要不稳定、不可交易或副本内限定等约束
```

积分制控制总强度，但不能单独解决平衡问题。

### 3. 独占组

强属性不只需要积分，还需要控制组合形状。

可以设计独占组：

```text
major_on_hit_effect:
  - 吸血
  - 变形
  - 放逐
  - 冻结
  - 即死
  - 精神控制

major_economy_effect:
  - 复制掉落
  - 复制自身
  - 增加金币产出

major_world_effect:
  - 重置区域
  - 传送房间
  - 改写出口
```

日常掉落可以牺牲上限以换取稳定：

```text
普通掉落最多 0-1 个 major 属性
普通副本非特殊产出最多 1 个 major 属性
特殊掉落可以突破
手工神器由设计者负责
```

### 4. 风险标签和约束需求

某些效果不是单独强或弱，而是高度依赖组合上下文。

例如复制、自我增殖、成本绕过、高频触发、全局范围、永久控制，都可能形成危险闭环。

effect 应声明风险形状：

```text
复制自身:
  risks:
    - self_replication
    - economy_breaker
    - persistence_amplifier
  requires_container:
    any_of:
      - limited_charges
      - consumes_self
      - non_persistent
      - non_tradeable

即死:
  risks:
    - encounter_skip
  requires_container:
    any_of:
      - low_chance
      - limited_charges
      - suppress_reward
      - self_risk
```

强效果不一定危险，弱效果也不一定安全。真正危险的是闭环：

```text
触发 -> 产生资源 -> 资源支持再次触发
触发 -> 复制触发源 -> 更多触发源 -> 更多触发
触发 -> 绕过代价 -> 高频触发 -> 无限收益
```

原则：

> 危险效果必须配约束效果。约束效果本身可能很强。弱效果如果参与闭环，也可能极危险。

### 5. 生成后审计

随机生成应该是：

```text
选择模板
选择预算和来源等级
随机 tag/effect/modifier
检查独占组
检查风险闭环
检查是否缺少约束
检查是否污染持久世界
通过则生成
不通过则 reroll、降级或添加副作用
```

这不是内容特例，而是通用安全机制。

## 手工神器与日常随机的边界

日常随机生成应该保守，牺牲理论上限以换取稳定和可信。

真正越界的物品应该来自：

```text
设计者手工神器
Boss 特殊产出
限时副本
异常事件
不可长期污染世界的生成机制
```

手工神器可以突破模板、积分和独占组，但必须仍然落到通用机制里，并显式声明风险和约束。

原则：

> 组合系统负责表达能力，设计者负责定义可出现的空间。

## 持久世界中的额外风险

MUD 不是单局 roguelike。随机强物品的风险不是只影响一局，而是可能长期存在并进入社会系统。

因此随机物品必须考虑：

```text
是否可交易
是否可复制
是否可修理
是否可继承
是否可长期保存
是否能影响其他玩家
是否能影响经济
是否能影响世界地形或内容
```

一次性卷轴可以很强，永久可交易装备要谨慎；副本内临时道具可以疯狂，持久世界掉落要保守。

## 第一版不实现这些复杂系统

这些讨论是底层抽象的压力测试，不是第一版功能需求。

第一版不需要实现：

```text
复杂合成
贴附槽位
随机神器生成
完整 trigger/effect DSL
脚本系统
风险审计器
复杂可见性规则
完整 tag 冲突系统
```

第一版真正需要保留的是：

```text
EntityID
Tags
少量 Components
Relations
Provider / EffectiveView 的设计空间
Command -> Event 的方向
持久化 live state
```

不要在第一版做会阻断未来的决定：

```text
Room 只是不可引用 struct
Exit 只是 map[direction]RoomID
Container 是 Item 子类
Player 和 NPC 是完全不同的世界对象
命令大量 switch 具体类型
tag 只是无语义字符串
物品效果写死在物品 ID 分支里
```

第一版可以很小，但底层要允许这些复杂系统以后自然长出来。

## 声明式语义与热加载脚本

Tag / Component 不应该变成脚本函数。它们是引擎可理解、可校验、可索引、可持久化的结构化语义。

```text
Tag / Component:
  portable
  container
  exit
  light_source
  flammable
  durability
  binding
```

这些事实必须在加载或热加载时被 ContentCompiler 静态检查，并编译成 flags、typed slots、provider lists 和 trigger index。命令系统、观察系统、持久化系统和清理系统不能靠“运行脚本问这个物体有什么 tag”来理解世界。

但脚本语言的核心价值也必须承认：**免 Go 编译的内容迭代**。Go 缺少适合这个项目的通用动态链接路径，不能让内容作者为了改一个房间行为、特殊物品反应或任务机关就完整重新编译服务端。

因此边界应是：

```text
声明式内容:
  定义实体有什么能力、状态、关系、展示、基础 requirement

热加载脚本 / DSL:
  定义少量需要频繁迭代的行为、trigger reaction、特殊 resolver、房间机关、NPC 对话或任务步骤

Go 引擎:
  提供稳定、受控、类型化的世界 API
  执行事务、持久化、权限、调度和安全边界
```

脚本是行为扩展点，不是实体语义的根基。

更准确地说，脚本层应是行为组合层，而不是第二套世界引擎。Lua 可以用有限的 hook、condition、selector、probability 和 effect primitive 拼装新行为，类似把已有积木组合成新道具或新机关。

```text
Hook:
  on_use
  on_enter
  on_pickup

Condition:
  has_tag
  state_equals
  chance

Selector:
  actor
  self
  room

Effect:
  emit
  set_state
  move_item
```

`move_item` 虽然会改变世界关系，但它是必要 primitive。很多 MUD 行为不是简单改数值，而是改变物体关系位置。例如：给武器涂油后，武器可以被点燃，但同时变滑，可能从手中滑落到地面。

```text
oil_on_weapon:
  condition:
    target has tag weapon

  effects:
    set_state(target, "oiled", true)
    set_state(target, "flammable", true)
    chance(20%) -> move_item(target, actor.equipment.hand, room.ground)
    emit("武器变得油亮而滑腻。")
```

这种行为必须能由内容热加载组合出来，否则每个特殊物品互动都要回到 Go 里硬编码。

但 `move_item` 必须是受控 effect request，而不是 Lua 直接改 relation。第一版可以只支持最小范围：

```text
allowed first-version move_item:
  room ground -> actor inventory
  actor inventory -> room ground

future move_item cases:
  equipment slot -> room ground
  inventory -> container
  container -> room ground
  actor hand -> room ground
```

禁止第一版脚本做：

```text
远程房间移动
移动其他玩家背包物品
批量移动
凭空创建或销毁物品
绕过绑定、重量、容量、权限和持久化规则
```

原则：

> `move_item` 是必要行为词汇，但只能作为 Go 验证的 effect request。Lua 可以请求改变物品关系，不能直接修改 relation graph。

日常内容迭代应该只组合这些已有词汇，因此不需要重新编译 Go：

```text
new item behavior:
  existing hook + existing conditions + existing effects
  hot reload content snapshot

new room mechanism:
  existing hook + existing selectors + existing effects
  hot reload content snapshot
```

如果要引入全新的 hook、condition 类型、selector 类型或 effect primitive，则属于扩展引擎词汇表，需要修改 Go 并重新编译。这个代价可以接受，因为随着游戏开发推进，基础词汇表会逐渐稳定，新增 primitive 的频率应该越来越低。

```text
content composition:
  frequent
  hot reload
  no Go rebuild

engine vocabulary expansion:
  less frequent
  Go code change
  rebuild required
```

原则：

> 允许内容免编译组合行为；不要求引擎免编译获得全新基础动词。热迭代服务日常内容生产，Go 重编译服务少量引擎词汇表扩展。

脚本形态不应限制为纯数据表。纯数据表便于校验，但表达力很快不够：NPC 对话、房间谜题、仪式机关、多个候选目标和运行时分支都会把数据 DSL 推向复杂自制语言。

第一版应允许两种行为定义形态：

```text
structured behavior table:
  适合简单行为
  易校验、易报错、易可视化

Lua function returning effect requests:
  适合需要 if/else、局部变量、循环、运行时查询和复用函数的行为
  仍不能直接修改世界
```

Lua 可以拥有普通编程语言的控制流：

```lua
function on_use(ctx)
  if not conditions.has_tag(ctx.actor, "can_read_runes") then
    return {
      effects.emit("你看不懂这些符文。"),
    }
  end

  if conditions.chance(ctx, 30) then
    return {
      effects.emit("石门缓缓打开。"),
      effects.set_state(ctx.self, "open", true),
    }
  end

  return {
    effects.emit("符文闪了一下，然后归于沉寂。"),
  }
end
```

但 Lua 不能拥有直接世界写权限：

```text
不能直接改 HP / state / inventory
不能直接移动 entity
不能直接写数据库
不能绕过 cooldown / transaction / permission
不能访问全局 world map 做任意修改
```

Go 执行器负责验证 Lua 返回的 effect requests，并通过 `WorldMutation` 提交。

因为 Lua 函数允许控制流，每次 hook 调用必须有执行预算。预算不是性能优化，而是脚本安全边界。

```text
budget dimensions:
  instruction count
  wall time
  returned effect count
  selector / query count
  stack / recursion depth
```

第一版按 hook kind 使用固定预算：`on_use`、`on_enter` 等即时 hook 预算很小；未来 `scheduled_event`、`world_event` 可以有更大预算。内容脚本第一版不自行声明预算等级，后续若允许 `small / medium / large`，大预算必须经过巫师或管理员审查。

超预算等同脚本失败：不提交 effect，不修改世界，走脚本错误码流程。

随机数也属于 Go 管理的执行上下文，而不是 Lua 自己的全局能力。脚本不能使用 `math.random()`、不能按当前时间自建 RNG，也不能循环抽随机直到成功。

```text
recommended:
  conditions.chance(ctx, percent)

forbidden:
  math.random()
  script-local clock-seeded RNG
  repeat random until success
```

每次 hook 调用由 Go 创建事件局部 RNG：

```text
RngContext:
  domain: script
  event_id
  script_id
  hook
  actor_id?
  entity_id?
  draw_count
```

不要让所有系统共享一个巨大随机序列。全局序列会让脚本、掉落、战斗、生成互相污染：某个脚本多消耗一次随机数，可能改变其他玩家后续掉落或战斗结果。

原则：

> Randomness is event-local, domain-separated, Go-owned, budgeted, and traceable. Lua can ask for chance; Lua cannot own randomness.

原则：

> Lua may use control flow, but every hook execution has a budget. Only Go applies world changes. Scripts compose and return effect requests, not mutations.

坏方向：

```text
tags(entity) -> lua table
can_pickup(entity, actor) -> arbitrary script everywhere
room description entirely generated by script
```

好方向：

```text
tags:
  portable: {}
  light_source:
    radius: 2

triggers:
  on_use:
    script: scripts/items/old_lantern.lua:on_use
```

原则：

> Tags are declarative, class-like semantic objects. Behaviors may be scripted and hot-loaded. Lua or another script runtime is an iteration tool, not a replacement for engine-understood tag semantics.

## 当前设计承诺

这个项目的抽象承诺是：

> 世界里没有“特殊代码对象”，只有“罕见规则组合”。

> tag 是玩家可学习、可依赖的语义单位；属性和效果可以组合，但不能欺骗玩家。

> 实体通过能力、关系、provider 和 effective view 参与规则，而不是通过继承树或硬编码类型分支。

> 手工内容、随机内容、合成内容和贴附内容最终都应该落到同一套 Entity + Strong Tag + Relation + Trigger + Effect 体系里。

这些原则比任何具体玩法都重要。
