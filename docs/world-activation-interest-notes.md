# 世界激活与兴趣范围设计笔记

MUD 的天然优势是空间离散、观察范围小。玩家通常只需要知道当前房间和少量相邻房间；服务器也不应该把整个世界当成持续运行的实时模拟。

核心原则：

> 不仅不向客户端推送全世界，大多数非活跃区域也不需要实时计算。世界应按兴趣范围、活跃集合、显式事件和按需投影运行。

## 客户端兴趣范围

客户端常驻 UI 只需要一个很小的 local interest scope。

```text
primary:
  当前房间

nearby:
  当前地图显示需要的相邻房间摘要

extended:
  特殊可见、可听、已探索或地图知识范围
```

典型规模：

```text
3x3 地图:
  当前房间 + 周围 8 格

普通出口节点:
  当前房间 + 若干出口

复杂交通/传送节点:
  可能有 20-30 个候选出口，但这些应作为候选列表或 routes 显示，不作为完整实时房间同步
```

## 不同范围同步不同信息

当前房间和相邻房间不应同步同样详细的信息。

```text
current room:
  完整 ObservationPayload
  可见实体
  可交互物
  visible exits / transition candidates
  氛围、危险、环境提示

nearby rooms:
  MapCellSummary
  名称 / 短标签
  known / unknown
  reachable / blocked
  connection mask
  danger hint
  强信号，例如战斗声、火光、烟雾、人声

farther known map:
  已探索或地图知识
  不同步实时细节
```

3x3 地图里的每格应是摘要，而不是房间完整状态。

```text
MapCellSummary:
  place_id
  label
  known
  reachable
  danger_level?
  marker?
  connection_mask
```

## 同步策略

WebSocket 允许主动推送，但推送范围应限于当前 session 的 active scopes。

```text
Always subscribed:
  status
  location
  environment summary
  current_room_observation
  map_3x3_summary
  log
  prompt / candidates

On-demand subscribed:
  inventory_detail
  equipment_detail
  quest_log
  skill_list
  map_detail
```

移动时：

```text
room.snapshot(new current room)
map_3x3.snapshot(new local map)
environment.patch if changed
status.patch if needed
log.event movement result
```

站着不动时，只推 dirty patch：

```text
同房间有人进出
当前房间物品变化
HP / 状态变化
环境变化
log event
```

## Patch 与 Snapshot

状态同步以 dirty patch 为主，低频 snapshot 兜底。

```text
dirty patch:
  变化后合并发送
  50-100ms flush cadence

periodic snapshot:
  低频全量校准
  常驻小状态可 30-60 秒一次
  大对象只在打开期间低频校准

version:
  每个 scope 有版本号
  patch 带 base_version 和 version
  客户端发现版本不匹配时请求 snapshot
```

LOG 与状态不同。LOG 是事件流，状态是当前值。

```text
log.event:
  狼咬了你一口。

status.patch:
  hp: 82 -> 71
```

## 世界激活

服务器也不应该默认实时计算全世界。大多数房间、NPC、物品、环境状态在无人关注且没有重要事件时可以不运行。

世界运行可分为：

```text
active:
  有玩家在场
  当前房间或强相关相邻房间
  正在战斗
  有重要 NPC 或特殊场景
  有到期 ScheduledEvent

sleeping:
  没有玩家
  没有重要 NPC
  没有到期事件
  不需要实时模拟

projected:
  长期状态只保存 last_projected_time 和参数
  被观察、进入、唤醒时按时间差结算
```

原则：

> 非重要 NPC 和普通区域，在没有玩家、没有重要事件、没有特殊场景时，不需要实时 AI、移动、战斗或环境计算。

## 激活触发

区域或房间可以被这些事件激活：

```text
玩家进入
玩家观察或远程侦察
ScheduledEvent 到期
重要 NPC 行动
世界事件影响
跨区域 Transition 即将发生
管理员或调试工具请求
```

激活时应先做 lazy projection：

```text
load / locate live state
calculate elapsed since last_projected_time
apply growth / recovery / decay / reset rules
create active room/NPC/combat state if needed
```

## 重要 NPC 与特殊场景

并非所有 NPC 都一样。

```text
普通野怪:
  无玩家时不需要实时行动
  进入区域时按 spawn/reset/projection 生成或恢复

重要 NPC:
  可以拥有 ScheduledEvent 或 world-level behavior
  即使无人关注，也可能需要推进部分状态

特殊场景:
  世界事件、战争、商队、拍卖、剧情仪式等
  可通过 durable job 或 explicit scheduler 运行
```

也就是说，特殊内容通过显式事件和激活规则运行，而不是因为全世界每 tick 都被扫描。

## 与时间系统的关系

这与时间系统的结论一致：

```text
tick:
  推进被动世界时间和 active systems

Scheduler:
  唤醒未来事件

Active Systems:
  只处理活跃集合

Lazy Projection:
  处理长期和离线变化
```

不要写：

```text
for every room in world:
  room.tick()

for every npc in world:
  npc.ai_tick()
```

而应写成：

```text
for active combat in activeCombatList:
  combat.tick()

for active npc in activeMobList:
  npc.tick()

for due event in Scheduler:
  event.execute_if_still_valid()
```

## 当前结论

```text
客户端只同步 local interest scope。
当前房间同步完整 ObservationPayload。
相邻房间同步 MapCellSummary。
复杂交通/传送节点通过候选列表表达。
服务器不实时计算全世界。
普通非活跃房间和 NPC 默认 sleeping。
长期状态通过 lazy projection 按需结算。
重要 NPC 和特殊场景通过显式 ScheduledEvent / durable job 运行。
```

后续还需要讨论：

```text
active room 的生命周期
active NPC 的唤醒和休眠规则
相邻房间是否需要实时监听强信号
MapCellSummary 的字段
重要 NPC 如何声明 world-level behavior
projection 与 area reset 的具体关系
```
