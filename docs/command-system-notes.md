# 命令系统设计笔记

这份笔记记录命令系统目前达成的框架层共识。它不决定具体游戏内容，也不规定未来每个特殊能力应该叫什么命令。它只描述玩家输入如何进入世界模型，以及命令、动作、对象、tag、ability、requirement、effect 之间的边界。

## 核心结论

命令不应该直接调用硬编码函数。玩家输入应该先被解析成系统语义上的 `ActionAttempt`，再由参与这个动作的实体、tag、capability、relation、环境共同决定它能否发生、如何发生、产生什么结果。

原则：

> 命令是文本入口，Action 是系统语义，Entity 是动作对象，Role 描述对象在动作中的位置，Tag/Capability 决定对象如何参与动作，Event/Effect 修改世界，Presentation 生成反馈。

## 命令执行管线

推荐的概念管线是：

```text
1. Parse
   文本 -> action name + noun phrases

2. Resolve
   noun phrases -> candidate entities
   处理别名、可见性、距离、容器、歧义

3. Bind Roles
   把对象绑定到 action roles
   例如 target/tool/host/attachment/source

4. Check Requirements
   actor、对象、环境、tag、relation 共同提供前置条件

5. Build Effects
   action 和参与对象共同生成 effects/events

6. Apply
   world loop 应用 events，修改 live state

7. Present
   给 actor、旁观者、目标生成不同反馈
```

命令系统不应该在 parser 阶段直接修改世界。parser 只负责把文本变成可解释的动作尝试。

文本解析可以在 session 层或输入层先做，但实际 entity resolve、requirement check 和执行必须在 world loop 中基于权威 live state 完成。验证和执行必须属于同一个原子流程。

第一版实现不需要建立完整复杂的 command framework，但必须保留这条骨架边界：

```text
parse
  raw input -> ParsedCommand / ActionAttempt hint

resolve
  当前 world state 下解析目标和 role

validate
  检查 actor、target、relation、tag requirements

mutate
  通过 WorldMutation / PositionTransaction 修改世界

present
  生成结构化 PresentationEvent
```

第一版可以直接实现 `look/go/say/get/drop/inventory/quit`，但不要让任何命令绕过以下入口：

```text
World loop:
  世界状态唯一写入口

PositionTransaction:
  relation 改动唯一写入口

PresentationEvent:
  玩家输出唯一入口

PersistStore:
  玩家长期事实唯一写入口
```

例如玩家输入：

```text
get apple
```

输入层可以先解析出：

```text
intent: take
noun phrase: apple
```

但不能在 session 层就假定某个 `apple_001` 一定可拿。等命令进入 world loop 后，系统需要重新基于当前状态解析或验证目标：

```text
当前 Place 是否仍有可见 apple？
apple 是否仍可触达？
apple 是否已被别人拿走？
目标容器是否仍有容量？
```

只有这些检查在 world loop 中通过，才提交对应的 action / transaction。

原则：

> 解析可以提前，权威 resolve、验证和执行必须在 world loop 中原子完成。任何入队前得到的对象候选都只是提示，不是执行时的事实。

## ActionAttempt

一个动作尝试应包含：

```text
ActionAttempt:
  action: unlock / attach / traverse / attack / invoke_ability / ...
  actor: EntityID
  roles:
    target: EntityID?
    tool: EntityID?
    source: EntityID?
    host: EntityID?
    attachment: EntityID?
    ability: AbilityID?
  raw_input: 原始输入文本
  context: 位置、可见性、当前时间、session 等
```

同一个动作可以有多个动作对象。例如：

```text
unlock door with key:
  action: unlock
  actor: player
  roles:
    target: door
    tool: key

attach page to blood bottle:
  action: attach
  actor: player
  roles:
    attachment: page
    host: blood_bottle

attack goblin with sword:
  action: attack
  actor: player
  roles:
    target: goblin
    tool: sword
```

这样命令系统不会被单对象动作锁死。

## Action 是主线，Tag 是参与者

不是由某个 tag “抢到动作执行权”并独自执行整个命令。更合适的模型是：Action 定义动作语义，参与对象的 tag/capability/relation 提供响应、约束、修饰和效果。

tag 在动作中的职责可以分为：

```text
Capability Provider
  说明对象能参与某动作
  例如 Exit 支持 traverse，Lockable 支持 lock/unlock

Requirement Provider
  提供动作前置条件
  例如 locked 要求 unlocked，narrow 要求 crawl 或 small body

Modifier
  修改动作参数或结果
  例如 slippery 让抓取更难，heavy 增加体力消耗

Effect Provider
  在动作成功时产生额外效果
  例如 cursed 在使用时反噬，flaming 在攻击命中时点燃

Reaction
  在动作发生后响应
  例如 fragile 被攻击后破碎，alarm 触发警报
```

这让 tag 能参与动作，但不让命令系统变成一堆目标对象的特殊分支。

## 出口动作示例

`go north`、`enter portal`、`crawl into hole` 可以收束成同一种底层动作：

```text
action: traverse
actor: player
roles:
  target: exit_entity
mode: walk / enter / crawl
```

`north`、`door`、`hole`、`portal` 都可以解析到具备 `Exit` capability 的 entity。不同输入主要影响 `mode`、需求检查和展示文本，而不是产生完全不同的底层系统。

例如：

```text
go north:
  target = north_exit
  mode = walk

crawl into hole:
  target = dog_hole_exit
  mode = crawl
```

`dog_hole_exit` 可以通过 requirement 要求 `mode = crawl` 或 actor 体型足够小。

## Ability 不注册新全局动词

特殊能力不应该通过不断新增全局动词来扩展命令系统。更干净的模型是：tag/capability 提供 `Ability`，Ability 声明自己可以被哪些通用动词入口触发、需要哪些 role、产生哪些 effect。

原则：

> 特殊能力不增加新的系统动词，只增加新的可调用能力。

Ability 的概念结构：

```text
AbilityDefinition:
  id
  display_name
  provided_by
  accepted_verbs
  role_schema
  requirements
  effects
  presentation
```

例如：

```text
ability: sword_blast
provided_by: 爆裂魔法 tag
accepted_verbs:
  - use
  - release
  - invoke
role_schema:
  source: sword
  target: visible enemy
requirements:
  - source is wielded
  - cooldown ready
effects:
  - explosive_damage(target)
```

不同文本输入可以归一为同一个底层动作：

```text
release sword blast at goblin
use sword blast on goblin
invoke blast from sword at goblin

=>

ActionAttempt:
  action: invoke_ability
  actor: player
  roles:
    source: sword
    ability: sword_blast
    target: goblin
```

## 通用动词与参数化行为

特殊能力的差异应优先通过参数和数据表达，而不是新增函数或新增系统动作。

不推荐：

```text
castSwordBlast()
prayAmuletBlessing()
activateRuneDoor()
whisperRingSecret()
```

推荐：

```text
invokeAbility(actor, ability, roles, context)
```

差异来自数据：

```text
ability_id
accepted_verbs
role_schema
requirements
effects
presentation
```

原则：

> 行为差异优先数据化，而不是函数化。特殊能力通过 role 参数、requirement、effect 和 presentation 改变通用 action 的行为；只有当动作语义本身不同，才新增 core action。

## 文本输出与 Presentation 边界

框架层不应该直接生成面向玩家的自然语言文本。所有玩家可见输出都应先生成结构化 presentation payload，再由 Presentation/i18n 层渲染成具体文本。

这套边界在 `presentation-system-notes.md` 中单独展开。命令系统、Transition、死亡结算、观察结果都应共享同一套结构化输出原则。

这不只适用于失败反馈，也适用于所有文本：

```text
命令成功反馈
命令失败反馈
look / examine / appraise 输出
房间描述
物品描述
战斗消息
旁观者广播
系统提示
调试/专家模式输出
```

原因是：同一个世界事实，对不同玩家可能需要不同表达。

差异来源包括：

```text
语言 / locale
观察者视角
技能等级
属性或感知能力
种族、文化或阵营风味
是否处于简短模式 / 详细模式
客户端是否需要结构化输出
```

例如同一个失败原因可以先表示为：

```text
Failure:
  code: ContainerRejectsItem
  container: quiver_001
  item: apple_001
  required: arrow_like
```

然后不同 presentation 层可以渲染为：

```text
普通中文:
  箭袋装不下苹果。

详细中文:
  这个箭袋只能容纳箭矢一类的细长弹药，装不下苹果。

英文:
  The quiver cannot hold the apple.

结构化客户端:
  {"code":"ContainerRejectsItem","container":"quiver_001","item":"apple_001"}
```

成功事件也一样。Action 或 Effect 不应该直接写：

```text
你拿起了苹果。
```

而应产生结构化结果：

```text
Event:
  type: EntityPositionChanged
  entity: apple_001
  from: located_in(room_001)
  to: contained_in(backpack_001)
  actor: player_001
  reason: take
```

Presentation 再根据接收者渲染：

```text
actor 看到:
  你把苹果放进了背包。

旁观者看到:
  某人收起了一个苹果。
```

原则：

> Action、Requirement、Effect、Event 只产生结构化事实和结构化原因；自然语言文本只在 Presentation/i18n 层生成。框架层禁止散落硬编码玩家文本。

这也是国际化的基础。硬编码文本会让后续多语言、风味文本、技能差异描述、客户端结构化输出都变成灾难。

## 解析优先级与歧义处理

命令解析应区分两层：

```text
语言解析层:
  这句话可能是什么意思？

世界解析层:
  在当前权威 live state 中，它实际能指向什么？
```

例如：

```text
open north
open door
open wooden door
```

语言解析层只需要得到：

```text
action: open
target phrase: north / door / wooden door
```

world loop 中的世界解析层再根据当前 Place、可见实体、装备、直接可访问容器和显式容器范围寻找候选。

默认搜索范围应遵守可访问性规则：

```text
显式指定对象优先:
  get pear from old bag
  use key on door

当前 Place 中可见交互对象:
  door, north, chest, goblin

玩家装备物品:
  worn / wielded / equipped items

装备中容器的一层直接内容:
  backpack contents, pockets, quiver contents

未装备容器内容:
  不进入隐式范围，必须显式 from / in / on
```

短命令只有在候选唯一、低风险且符合上下文时才应自动解析。一旦存在多个合理候选，系统不应该猜。

歧义应返回结构化 payload：

```text
AmbiguousTarget:
  phrase: sword
  action: wield
  candidates:
    - rusty_sword
    - silver_sword
```

Presentation 层可以把它渲染为：

```text
你指的是哪一把剑？
1. 锈裂短剑
2. 银质仪式短剑
```

框架应支持下一步编号选择候选，减少玩家因为一时疏忽反复输入完整命令的垃圾时间。编号选择仍然必须在 world loop 中重新验证候选是否仍然有效；如果候选已经移动、消失或不可触达，应返回新的结构化失败原因。

候选列表只对下一条输入有效。

```text
上一条输出候选列表：
  1. 锈裂短剑
  2. 银质仪式短剑

下一条输入是合法候选值：
  1
  -> 作为候选选择处理，并重新验证目标

下一条输入不是合法候选值：
  look
  -> 丢弃候选列表，按普通命令解析 look
```

候选列表不跨命令长期保留。这样可以避免玩家过了一段时间输入编号，却引用已经过期或世界状态已经改变的旧候选。

原则：

> 歧义不自动猜。框架返回结构化候选列表，Presentation 负责显示，后续输入可以选择候选，但执行前仍需基于权威状态重新验证。

第一版可以先做到“结构化歧义 + 提示更具体地输入”；编号候选选择可以作为后续增强。但数据结构应从一开始支持候选列表。

## 命令别名归属

命令别名至少分三类，不应该混在 Go 逻辑里：

```text
语言包别名:
  get/take, look/l, inventory/i
  属于 locale / command grammar

实体别名:
  north/n/door/wooden door
  属于 entity display 或 content data

能力别名:
  blast/release/cast
  属于 AbilityDefinition 的语义能力，但具体文本仍由语言包表达
```

规则系统应使用语义 ID：

```text
ActionID
EntityID
AbilityID
TagID
```

玩家输入的具体词属于解析和语言层。这样才能支持 i18n、别名、玩家自定义缩写、不同客户端显示，以及同一个能力在不同语言中的不同表达。

## 同名 Ability 的短命令策略

理想情况下，同名 ability 不应造成歧义；但框架不能假设内容永远没有歧义。由于贴附、装备、状态和特殊物品都可能提供 ability，同一个角色可能同时拥有多个同名能力。

例如，某个能力既可能来自手中武器，也可能来自鞋子、戒指、披风或其它装备。框架不应禁止非手部装备提供主动能力，否则会破坏直觉和组合性。

短命令可以作为语法糖，但必须保守：

```text
blast goblin
```

只有在能根据上下文安全确定来源时，才展开为：

```text
invoke_ability:
  ability: blast
  source: resolved_source
  target: goblin
```

默认策略：

```text
1. 显式来源永远优先。
   blast goblin with silver_shoe
   release blast from ring

2. 如果主手装备提供该 ability，短命令默认使用主手来源。
   这符合玩家直觉：即使鞋子也能释放 blast，手中拿着冲锋枪时，默认应使用手里的武器。

3. 如果主手没有提供该 ability，而多个其它装备提供同名 ability，则短命令不猜，返回 AmbiguousAbility。

4. 如果主手没有提供该 ability，且只有一个其它装备提供，则可以解析到该唯一来源。

5. 所有解析结果在 world loop 执行前仍需重新验证。
```

这条规则承认手部槽位在默认工具选择中有交互便利上的优先级，但不把手部设计成架构上独有的动作来源。

原则：

> 任意装备都可以提供 ability；显式来源优先；短命令优先主手来源；没有主手来源且存在多个候选时必须消歧。

## 万能 do 的定位

底层可以存在通用 ability invocation，但玩家层不应鼓励万能 `do`。

如果 `do` 过于方便，玩家会把它当成唯一命令，绕过更自然的世界表达。因此 `do` 应该只是：

```text
调试入口
测试入口
无障碍或专家兜底
歧义严重时的显式指定方式
```

默认玩家引导应使用 ability 声明的自然表达，例如：

```text
release sword blast
activate rune
pray amulet
invoke sigil
```

而不是：

```text
do sword blast
```

框架层结论是：公开命令可以有多个通用动词入口，但底层应尽量归一到少数 action，例如 `invoke_ability`。具体哪些表达更符合某个能力，属于内容和呈现设计，不在框架阶段定死。

## 框架需要准备什么

框架层应准备：

```text
Action 注册表
通用动词入口
AbilityDefinition 数据结构
RoleSchema
Requirement 管线
Effect/Event 生成管线
Entity resolver
歧义处理接口
Presentation 接口
结构化输出 payload
i18n/message catalog 边界
AmbiguousTarget / AmbiguousAbility payload
候选列表后续选择机制
```

框架层不应该现在决定：

```text
每个特殊能力叫什么
哪些自然动词最适合某个神器
玩家界面是否显示 do
具体技能/装备/神器能力如何写
```

## 仍待讨论的问题

后续还需要继续讨论：

```text
候选列表是否需要支持 more/next page？
歧义候选是否需要按风险或常用程度排序？
```

这些问题留到命令系统后续讨论中继续展开。
