# Spawn 与生命周期设计笔记

运行时生成物不能只用“它现在在哪个房间”来判断语义。一个物品、NPC 或临时实体是否持久、是否会被 reset 清理、是否能转化为玩家权益，取决于它的来源和生命周期策略。

核心原则：

> SpawnOrigin 回答“它为什么出现”；LifecyclePolicy 回答“它什么时候消失、什么时候持久化、什么时候转化为玩家利益”。

## SpawnOrigin

SpawnOrigin 记录实体的来源。

```text
template_spawn:
  内容定义或区域 reset 生成

npc_drop:
  NPC 死亡或掉落表生成

admin_spawn:
  管理员、调试或补偿手动生成

quest_spawn:
  任务、剧情、仪式或机关生成

world_event:
  世界事件生成

dungeon_run_local:
  副本 run 内生成

player_created:
  玩家制作、建造、放置或召唤生成
```

来源只解释“为什么出现”，不单独决定持久化。

## LifecyclePolicy

LifecyclePolicy 记录实体的生命周期语义。

```text
volatile:
  临时 live state
  可因重启、reset、清理而消失

reset_managed:
  由区域/房间 reset 管理
  缺失时可补刷，过量或过期时可清理

run_local:
  只属于某个副本、挑战、梦境或特殊实例
  rollback / settle / instance end 时清理

player_interest_on_pickup:
  在地上时是临时状态
  玩家拾取成功后进入玩家权益边界并持久化

player_dropped_ground_item:
  玩家主动丢弃或从玩家持久物品转移到地面
  保留具体实例、词条、耐久和历史
  按地面清洁策略处理

death_dropped_ground_item:
  死亡惩罚导致的地面掉落
  保留具体实例，但清理、损坏、找回规则由危险度决定

persistent:
  具体实例需要跨重启保存
  位置、状态和重要属性进入数据库

quest_interest:
  与任务关键进度或玩家承诺绑定
  持久化规则由任务系统明确处理
```

## 常见组合

### 普通房间默认刷物

```text
origin: template_spawn
lifecycle: reset_managed
```

语义：

```text
房间 materialize 或 reset 时生成
被拿走后，reset 可按规则补刷
服务器重启可丢失地上实例
玩家拾取成功后转化为玩家权益
```

`reset_managed` 默认适合低价值、易获取、工具性或环境性物品。它不需要资源守恒；玩家因为服务器重启、区域 reset 或定时刷新获得多个同类实例是允许的。

```text
适合 reset_managed:
  树上的果子
  路边的野草
  前期区域山洞门口的灯笼
  临时帮助用火把
  训练区普通练习物品
```

原因：这类物品通常售价不高、用途有限、主动获取容易。为了阻止重复获取而追踪每个地面实例、每次 reset 消耗或每次重启前状态，复杂度不值得。

### 怪物掉落

```text
origin: npc_drop
lifecycle: player_interest_on_pickup
```

语义：

```text
地上时是临时掉落
没人拿可腐烂、reset 清理或重启消失
玩家拾取成功后持久化为玩家物品实例
```

### 副本内物品

```text
origin: dungeon_run_local
lifecycle: run_local
```

语义：

```text
只在当前 DungeonRun 内有效
不能直接带出副本
结算前属于 pending/run-local state
rollback 时清理
settle 时按副本规则转化为奖励或丢弃
```

### 管理员生成物

```text
origin: admin_spawn
lifecycle: explicit
```

管理员生成物不应有默认 lifecycle。

```text
debug 临时物:
  volatile

世界修复物:
  persistent 或 reset_managed

玩家补偿物:
  直接进入 player interest，或 persistent item instance
```

原因：管理员生成可能是调试，也可能是补偿，默认错了会造成事故。

### 唯一世界物品

```text
origin: world_event 或 content_initialization
lifecycle: persistent
```

语义：

```text
有具体 instance id
位置和状态进入数据库
重启后仍存在
不会被普通 reset 凭空再刷
```

### 玩家制作或建造物

```text
origin: player_created
lifecycle: persistent 或 player_interest
```

语义：

```text
通常属于玩家利益或长期世界变化
默认不应因重启消失
具体持久化范围由制作/建造系统定义
```

### 玩家主动丢弃物

```text
origin: player_created / player_interest / prior_persistent_item
lifecycle: player_dropped_ground_item
```

语义：

```text
物品保留具体 instance identity
词条、耐久、绑定信息、历史不丢
地上时可以被其他玩家拾取，除非物品规则禁止
由地面清洁策略决定何时清理、回收或进入失物招领
```

玩家主动丢弃不等于物品退化成普通 volatile 垃圾。BatMUD 一类游戏的魅力之一，就是玩家可能在世界中捡到其他人留下的独特物品。

### 死亡掉落物

```text
origin: death_penalty
lifecycle: death_dropped_ground_item
```

语义：

```text
死亡导致的掉落保留物品实例身份
能否找回、能否被他人拾取、多久清理，由区域危险度和死亡策略决定
D3 / D4 的规则可以显著更严厉
```

## SpawnService

运行时生成实体应通过 SpawnService 或等价服务完成。

```text
SpawnService.SpawnItem(
  definition_id,
  target_place,
  origin,
  lifecycle,
  reason,
)
```

SpawnService 负责：

```text
查找 RuntimeEntityTemplate
创建 LiveEntity / entity_id
初始化 mutable state
设置 SpawnOrigin 和 LifecyclePolicy
通过 PositionTransaction 放入目标位置
标记 observation dirty
必要时创建数据库 instance
生成结构化 PresentationEvent
```

临时给房间刷一把普通铁剑：

```text
definition_id: item.iron_sword
target_place: room.b
origin: admin_spawn 或 quest_spawn 或 npc_drop
lifecycle: volatile / player_interest_on_pickup
```

如果玩家拾取成功：

```text
InventoryService.PickUp(actor, item, room)
  -> 玩家权益确认点成立
  -> item instance 进入数据库
  -> location = actor.inventory
```

如果没人拿且服务器重启：

```text
volatile / player_interest_on_pickup 地上实例可以消失
```

## 与数据库的关系

是否进数据库不由“它在不在房间里”决定，而由 lifecycle 决定。

```text
地上普通掉落:
  不持久，不是玩家利益

地上唯一神器:
  持久，但未必属于玩家

玩家背包里的普通剑:
  持久，属于玩家利益

副本内拿到的临时剑:
  run_local，不属于外部玩家利益

玩家主动丢到地上的独特剑:
  仍是具体 item instance
  按 player_dropped_ground_item 规则清理或找回
```

原则：

> 物品是否持久，不由当前位置决定；物品是否属于玩家利益，不由它是否被生成决定。持久化和玩家权益都必须由 lifecycle 与确认点决定。

## 地面清洁策略

重启不是日常，不应把“服务器重启会不会丢”作为普通地面物设计的唯一中心。正常游玩中更重要的问题是：世界里会不会堆满垃圾，以及玩家能否遇到其他人留下的有趣物品。

地面清洁可以由清洁系统、NPC、区域规则或抽象回收机制表达。

推荐策略：

```text
低价值物品:
  固定时间后清理
  例如普通垃圾、廉价消耗品、低级白装

高价值物品:
  清理时间增加
  稀有度、词条数量、装备评分、玩家制作成本都可以提高保留时间

绑定装备 / 重要玩家物品:
  地上清理后不直接消失
  进入 lost_and_found / 失物招领
  玩家可通过费用、任务、NPC 或时间限制找回

危险等级 D3 / D4:
  所有装备都视作低价值固定时间清理
  不因绑定或高价值获得长时间地面保护
  是否进入失物招领由死亡/区域规则决定，默认更严厉
```

示例字段：

```text
GroundCleanupPolicy:
  value_tier
  dropped_at
  cleanup_after
  lost_and_found_eligible
  danger_override
  owner_hint?
  last_seen_at?
```

例子：普通低价值物品。

```text
wooden_club:
  cleanup_after: 15min
  lost_and_found_eligible: false
```

例子：高价值非绑定物品。

```text
rare_sword:
  cleanup_after: 2h
  lost_and_found_eligible: false
  can_be_looted_by_others: true
```

例子：绑定装备。

```text
bound_armor:
  cleanup_after: 30min
  on_cleanup: move_to_lost_and_found
```

例子：D3 / D4 区域。

```text
danger_level >= D3:
  treat_all_ground_equipment_as_low_value
  cleanup_after: fixed_short_duration
  lost_and_found_eligible: false unless explicitly protected
```

原则：

> 地面物品可以保留世界感，但必须有清洁策略。低价值固定时间清理，高价值延长，绑定装备可进入失物招领；D3/D4 覆盖为严厉清理规则。

## 副本空间

副本是完全独立的空间。副本内物品、怪物、机关、掉落和临时状态都属于 run-local state。

```text
不同玩家进入同一个副本入口:
  不一定进入同一个 run
  默认各自进入独立实例

组队进入:
  同一队伍成员可以进入同一个 run

离开、失败、rollback、服务器重启:
  副本内东西直接重置或丢弃
  不进入外部世界地面清洁系统
```

副本内获得物只有在副本结算时才可能转化为外部奖励。副本中途拿到的剑、药、钥匙、临时装备默认都是 `run_local`。

原则：

> 副本不是普通区域的一部分，而是隔离 run。副本内生命周期不和开放世界地面清洁混用。

## 与内容定义的关系

内容定义可以声明默认 spawn 规则。

```text
PlaceDefinition:
  spawns:
    - id: practice_sword_spawn
      definition: item.practice_sword
      origin: template_spawn
      lifecycle: reset_managed
      reset: if_missing
```

这表示“这个房间应该如何生成默认内容”，不是数据库中预先存在一个固定 item instance。

## 当前结论

```text
生成物必须记录 SpawnOrigin 和 LifecyclePolicy。
普通内容默认刷物使用 template_spawn + reset_managed。
reset_managed 低价值/工具性物品不追求资源守恒，允许玩家通过 reset、定时刷新或重启后重建获得多个实例。
怪物掉落使用 npc_drop + player_interest_on_pickup。
副本内物品使用 dungeon_run_local + run_local。
唯一世界物品使用 persistent。
管理员生成必须显式指定 lifecycle。
玩家拾取成功是普通地面物转化为玩家权益的确认点。
```

后续还需要讨论：

```text
玩家从背包丢弃已持久物品后，生命周期如何变化
尸体和掉落容器是否也是 spawn entity
reset_managed 如何避免重复刷物
persistent unique item 被删除或摧毁后的记录方式
```
