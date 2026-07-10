# 空间与关系模型设计笔记

这份笔记记录空间、位置、容器、装备和可访问范围的框架层共识。文中的例子只用于验证框架抽象是否能落地，不代表实际游戏内容。未来真实游戏可以完全不使用这些例子。

## 核心原则

空间系统的目标不是模拟真实物理，而是支撑一个符合直觉、可推理、可维护的文字游戏世界。

核心结论：

```text
房间是 Entity + Place。
玩家的位置必须指向 Place。
位置是唯一的，不能冲突。
所有权、绑定、来源、记忆都不是位置。
贴附后的对象不再作为独立位置实体存在。
容器必须有容量和接收规则。
装备中的容器内容进入玩家直接可访问范围。
未装备容器的内容需要显式访问。
```

## Place 与房间

框架中不应该用 `is_room` 作为底层分类。更合适的是 `Place` 能力。

```text
Room = Entity + Place
```

`Place` 表示该实体可以作为角色的位置基准：

```text
玩家可以位于其中。
look 默认观察当前 Place。
say 默认广播到当前 Place。
go/traverse 默认查找当前 Place 中的 Exit。
drop 默认把物品放到当前 Place。
spawn 通常把 NPC 或物品放入某个 Place。
```

内容层和作者工具仍然可以使用“房间”这个词，因为对作者和玩家来说它直观。但规则系统应该检查：

```text
has Place
```

而不是：

```text
is_room == true
```

## 普通房间拓扑优先

很多看似特殊的空间都可以用普通 Place、出口、传送和描述文本表达。例如“怪物体内”“船舱”“梦境”“袋中世界”等，不一定需要底层支持复杂嵌套空间。

优先使用：

```text
Place
Exit
Traverse / Teleport event
Description
Requirement
```

不要为了理论上的“房间也可能在实体里”提前设计复杂嵌套空间。只有当普通房间拓扑和描述文本确实不够时，再扩展。

## 位置唯一

一个普通可移动实体同一时间只能有一个位置。

推荐的位置类型：

```text
located_in Place
  位于房间、场景、地点中。

contained_in Container
  位于某个容器中。

equipped_by Actor
  被角色装备、穿戴、持握。
```

位置互斥。一个物体不能同时在地上、箱子里、背包里和玩家身上。

装备不是背包物品的布尔状态。被装备时，物品的位置关系已经从容器/背包转移到装备槽位；它不应该同时仍然存在于背包列表里。

```text
错误模型:
  sword contained_in backpack
  sword.equipped = true

正确模型:
  sword equipped_by actor slot right_hand
```

这样才能清楚回答“剑到底在哪里”。一把剑在你手上，就是在你手上，不是“背包里一个被装备的剑”。

动作会移动位置：

```text
take sword:
  located_in current_place -> contained_in accessible equipped container

drop sword:
  contained_in/equipped_by -> located_in current_place

wear cloak:
  contained_in accessible container -> equipped_by actor

put apple in chest:
  current position -> contained_in chest
```

## 非位置字段

这些不是位置：

```text
owned_by
bound_to
created_by
known_by
source_context
provenance
```

它们可以限制使用、交易、装备、拾取或显示，但不表示物品在哪里。

例如：

```text
owned_by player_a
bound_to player_a
located_in room_b
```

表示这件物品在房间里，但所有权或绑定属于 `player_a`。它不是因为 `owned_by` 就自动在玩家背包里。

## 贴附不是位置

贴附后的对象不再作为普通独立位置实体存在。它成为 host 的 attachment slot 内容或 attachment state。

例如：

```text
robe has liquid attachment: oil
robe has solid attachment: charm
```

贴附物不再同时拥有：

```text
located_in Place
contained_in Container
equipped_by Actor
```

如果玩家拆下贴附物，它才重新进入位置系统。

贴附通过 provider 改变 host 的 `EffectiveView`，而不是制造新的空间位置。

## 容器类型

容器是拥有容量和接收规则的能力，不是根类型。

每个容器至少需要：

```text
capacity_volume
accepted_rules
contents
access_policy
```

容器内物体的体积和不能超过容器容积。

### 固定容器

固定容器是场景设施，通常不可携带，默认位于 Place 中。

例子：

```text
宝箱
商店货架
仓库柜
尸体
祭坛
```

固定容器可以容纳普通物品，也可以容纳可携带容器。它本身通常不进入另一个容器。除非特殊调试或特殊内容明确需要，否则固定容器只应 `located_in Place`。

### 可携带容器

可携带容器可以被玩家携带、装备或放入固定容器。

默认规则：

```text
可携带容器不能容纳其他容器。
```

也就是“袋子不能装袋子”。这可以避免无限递归、计算量爆炸、数据量膨胀和玩家构造服务器压力。

可携带容器仍然可以通过 `accepted_rules` 限定自己的用途。例如背包可以接受普通固体物品，箭袋只接受弹药，药剂带只接受药瓶，裤子口袋只接受小型固体物品。规则重点不是“所有容器都一样”，而是“每个容器都明确声明自己能装什么”。

### 装备中的容器

装备在角色身上的容器，其直接内容进入玩家的默认可访问范围。

例如：

```text
背包 equipped_by player
苹果 contained_in 背包

裤子 equipped_by player
香蕉 contained_in 裤子口袋
```

玩家可以直接：

```text
eat 苹果
eat 香蕉
```

因为它们在装备中容器的一层直接内容里。

### 未装备容器

未装备的容器可以被携带，但它的内容不会自动进入直接可访问范围。

例如：

```text
旧袋子 contained_in 装备中的背包
梨 contained_in 旧袋子
```

玩家不能直接：

```text
eat 梨
```

必须显式访问：

```text
get 梨 from 旧袋子
eat 梨
```

这条规则防止隐式范围无限展开。

## 可访问范围

命令解析应区分隐式范围和显式范围。

### 隐式范围

隐式范围用于 `eat apple`、`wield sword`、`use potion` 这类命令。

它包括：

```text
当前 Place 中可见/可触达实体
actor 自身装备的物品
装备中可访问容器的一层直接内容
```

隐式范围不递归展开容器中的容器。

### 显式范围

显式范围用于：

```text
look in bag
get pear from bag
take coin from chest
```

只要指定的容器可触达、可打开、允许查看或取用，就可以访问其中内容。

## 背包、袋子与专用容器

背包应该是真实容器，而不是单纯属性加成。但装备后的背包不需要玩家手动打开管理，它的直接内容进入默认 inventory/命令范围。

角色可以携带多个背包或袋子，但装备槽位有限。只有装备中的容器提供直接访问。

专用容器通过 `accepted_rules` 限制内容：

```text
箭袋:
  接受箭、弩矢、飞镖等远程弹药
  不接受苹果

药剂带:
  接受小瓶药剂
  不接受剑

裤子口袋:
  接受小型固体物品
```

袋子的价值可以来自：

```text
external_volume 小
capacity_volume 大
内容物重量照常计算
内容物体积不按原样暴露给外部容器
```

也就是说，袋子“里面比外面大”，但袋中物品仍然增加总重量。

## 装备区与动作提供

装备区不应该被分成“手是特殊动作区，其他槽位是被动区”。更干净的模型是：**任何被装备的物品都可以通过 tag、component 或 ability provider 向角色提供新动作、修饰动作或改变可用选项。**

手部槽位仍然重要，但它的重要性来自内容设计和玩家直觉，而不是框架特例：大多数武器、工具、火把、法杖等需要装备在手上才能使用，因此手部装备常常提供最多主动动作。

框架层规则：

```text
装备物品可以提供能力。
能力是否可用，由装备状态、槽位、Requirement 和 AbilityDefinition 决定。
手只是普通装备槽位之一。
```

例如，以下情况都应该通过同一机制表达：

```text
手上的工具提供 cut / dig / pry 等能力。
脚上的靴子提供特殊移动能力。
披风提供隐藏或滑翔能力。
戒指提供 invoke / whisper 等能力。
背包提供可访问容器空间。
```

这些例子只是验证框架，不代表实际游戏内容。

因此，命令解析不应硬编码“只有手中物品能提供新动作”。它应该检查 actor 当前装备提供的 ability providers，再结合动作本身的 role schema 和 requirement 判断是否可用。

### 必须装备才能提供动作

普通物品放在容器里时，不应该默认提供主动动作。要让物品的新动作进入玩家可用动作集合，通常必须先装备到某个合适槽位。

```text
物品 contained_in 背包:
  可以被取出、查看、使用为显式对象。
  默认不向玩家提供新的隐式动作。

物品 equipped_by actor:
  可以通过 provider 向 actor 提供动作、修饰和能力。
```

这样可以避免背包里所有物品的能力同时污染命令解析，也让装备选择有意义。

因此，装备和背包必须通过互斥位置关系区分，而不是通过背包物品上的 `equipped` 标记区分。

### 装备置换

装备动作可以支持置换，不要求玩家先手动 remove 或 drop。

```text
wield new_sword:
  如果目标槽位已有 old_sword，系统尝试把 old_sword 移回可访问容器。
  如果没有可用容器，可以失败，或根据动作规则掉落到当前 Place。
  new_sword 再进入装备槽位。
```

置换仍然走统一位置变更规则：旧装备离开 `equipped_by`，新装备进入 `equipped_by`。框架只需要支持原子化的多步位置变更和失败回滚；具体失败反馈属于 Presentation。

## 体积与负重

容器检查体积：

```text
contents_volume_sum <= capacity_volume
```

玩家还检查总负重。

负重不一定阻止拾取。可以允许玩家先拿到东西，再通过移动限制产生压力：

```text
take treasure:
  如果有容器容积可放入，允许拾取。
  重新计算总重量。
  如果超重，进入 overloaded 状态。

go north:
  如果 overloaded，不能移动或受到严重惩罚。

drop / sort / put:
  仍然允许，以便玩家整理。
```

这符合“玩家可以先抢到宝贝，再慢慢扔掉垃圾”的体验。

## Vessel 与物质内容

液体、粉末、气体等非固体物质默认不必是独立 Entity。它们可以是 `SubstanceState`，存在于 Vessel、表面残留、地点残留或 attachment state 中。

例子：

```text
血瓶:
  Entity: bottle
  Physical: solid
  VesselContent:
    substance: blood
    phase: liquid
    amount: 200ml
    tags: [血液]
```

背包能装血瓶，是因为瓶子是固体实体；不是因为背包能装血液。

```text
put blood in bag:
  失败，需要能盛液体的容器。

put bottle of blood in bag:
  成功，因为 bottle 是 solid entity。
```

物态仍然有规则意义。液体、粉末、气体决定它们需要什么承载、如何转移、是否能直接放入某类容器、如何参与配方和动作。

## 效果传播边界

容器内容不应该被每个 tick 无限递归扫描。

默认原则：

```text
效果作用于直接目标。
是否影响内容物，由容器自身响应事件并显式释放或传播。
```

例如：

```text
火烧宝箱:
  先结算宝箱本体。

宝箱毁坏:
  释放直接内容物到当前 Place。

袋子掉出后:
  如果处在火焰环境中，再结算袋子。

袋子毁坏:
  再释放袋子内容物。
```

不要默认每秒递归引燃所有嵌套内容。

## 位置变更事务

改变位置是危险操作。`take`、`put`、`drop`、`wear`、`remove`、`wield`、`get from` 等命令表面不同，但底层都在把实体从一个位置移动到另一个位置。

这些操作应统一进入位置变更事务：

```text
PositionTransaction:
  actor
  changes:
    - entity A: old_position -> new_position
    - entity B: old_position -> new_position
  reason/action
```

例如装备置换：

```text
wield new_sword:
  old_sword equipped_by player -> contained_in accessible_container
  new_sword contained_in accessible_container -> equipped_by player
```

位置变更必须先完整校验，再一次性提交。

```text
validate:
  新装备是否能进入目标槽位
  旧装备是否能离开目标槽位
  旧装备是否有地方收纳
  容器容量是否足够
  容器 accepted_rules 是否允许
  绑定、锁定、任务状态是否允许移动
  变更后是否仍满足位置唯一

commit:
  所有变更一次性生效
```

如果任何一步失败，整个事务失败，世界状态保持不变。

原则：

> 不允许中间状态暴露给世界。失败时不自动把物品掉在地上，除非玩家显式 `drop`，或某个明确的游戏效果要求掉落。

这样可以避免玩家因为一时没记住背包容量，在置换装备时把主力装备掉到地上。

## World Loop、数据库事务与幂等

数据库事务可以保证持久化写入不写半截，但不能单独保证游戏世界状态不变半截。框架需要同时处理三层安全：

```text
world loop 事务:
  保证内存 live state 的原子性

数据库事务:
  保证持久化写入的原子性

command_id:
  保证网络重试不会把同一命令执行两次
```

推荐原则：

```text
所有 live state 修改只经过 world loop。
PositionTransaction 在 world loop 中先校验，再提交。
网络断线不取消已经进入 world loop 队列的命令。
每个客户端命令有唯一 command_id。
同一个 command_id 重试时返回同一个结果，不重复执行。
每个可移动实体只有一个当前位置。
数据库用事务保存事件和受影响状态。
持久化失败时进入保护模式，不继续接受会制造内存/数据库分叉的新命令。
```

命令生命周期可以理解为：

```text
客户端发送完整命令
  -> 服务器分配/接收 command_id
  -> 命令进入 world loop 队列
  -> world loop 校验并提交或拒绝
  -> 记录结果
  -> 返回给客户端
```

如果客户端在命令入队后断线，命令仍然完整处理或完整拒绝。玩家重连后可以看到结果或重新查询状态。

这三层共同防止：

```text
装备换到一半
网络重试复制物品
数据库写一半导致丢物品
内存状态和持久化状态长期分叉
```

## 框架需要准备什么

框架层应准备：

```text
Place component
Position union: located_in / contained_in / equipped_by
Container capability
Container accepted_rules
Container capacity_volume
Container access_policy
Equipped container direct-access scope
Explicit container access
VesselContent / SubstanceState
Weight and volume calculation hooks
Overloaded state or movement requirement
Attachment as host state, not independent position
PositionTransaction
command_id 幂等记录
world loop 原子提交边界
持久化失败保护模式
```

框架层不应该现在决定：

```text
具体背包容量
箭袋到底能装哪些弹药
瓶子能装多少血
哪些装备带口袋
哪些袋子稀有
具体负重公式
超重惩罚数值
```

这些属于内容设计或后续规则设计。

## 当前结论

空间和容器系统应采用：

```text
房间是 Place，不是硬编码特殊类。
位置唯一，只有 located_in / contained_in / equipped_by 等互斥状态。
所有权、绑定、来源不是位置。
贴附不是位置。
容器有容量和接收规则。
装备中的容器提供直接访问。
未装备容器需要显式访问。
可携带容器不能装容器。
非固体物质以 SubstanceState 存在，不必默认实体化。
效果传播不递归穿透容器。
```

这套规则牺牲了一些物理模拟完整性，但换来了玩家直觉、服务器安全、命令解析简单和内容可维护性。
