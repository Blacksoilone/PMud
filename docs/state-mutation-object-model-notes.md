# 状态修改与对象模型设计笔记

世界状态不能被任意系统直接改字段。我们要采用面向对象的领域模型：对象拥有自己的状态、不变量和行为；外部通过有语义的方法、服务或 mutation 请求改变状态。

核心原则：

> 不允许随意改字段。所有影响玩家权益、位置唯一性、可见世界事实、持久化、死亡、Transition、背包装备的变化，都必须通过受控领域行为或 WorldMutation。

## 为什么需要受控修改

如果每个系统都能直接改结构体，会很快失控：

```text
combat 直接改 hp
poison 直接改 hp
death 直接改 location
transition 直接改 location
inventory 直接改 item relation
presentation 读取半成品状态
persist 不知道哪些变化需要保存
```

结果是：

```text
不变量难维护
dirty patch 难生成
玩家权益边界不清楚
死亡和结算容易重复触发
状态变化难调试
未来 replay / rollback / audit 很难
```

## 对象拥有行为

领域对象不只是数据袋。它应暴露有意义的行为。

例如 Actor：

```text
Actor.ApplyDamage(source, amount, damage_type)
Actor.Heal(source, amount)
Actor.AddStatus(status, source)
Actor.RemoveStatus(status_id, reason)
Actor.CanMove(reason)
```

例如 Inventory：

```text
Inventory.CanAccept(item)
Inventory.AddItem(item, reason)
Inventory.RemoveItem(item_id, reason)
Inventory.FindUsableItem(key)
```

例如 Place：

```text
Place.CanEnter(actor, transition_context)
Place.OnEnter(actor, context)
Place.VisibleEntities(observer)
Place.ObservationFacets(observer, mode)
```

对象方法负责维护本对象的不变量：

```text
hp 不超过 max_hp
死亡状态不能重复进入
背包容量不能超限
装备槽不能同时装备两个主手武器
Place 只接受允许进入的 actor
```

## 跨对象变化走服务或 WorldMutation

对象方法适合维护单对象不变量，但很多操作跨多个对象。

例如：

```text
移动:
  actor location
  source place occupants
  destination place occupants
  visibility / observation

拾取物品:
  item location
  room contents
  inventory contents
  player interest persistence

交易:
  actor A inventory/money
  actor B inventory/money
  item ownership

副本结算:
  run state
  ticket reservation
  player rewards
  pending rewards
```

这类操作应由服务或 WorldMutation 统一处理。

```text
TransitionService.Execute(plan)
InventoryService.PickUp(actor, item, place)
TradeService.Exchange(a, b, offer)
DeathService.Resolve(actor, context)
DungeonService.Settle(run)
```

服务负责：

```text
协调多个对象
验证跨对象不变量
调用对象行为
提交 WorldMutation
记录 dirty state
生成结构化 PresentationEvent
触发持久化权益变更
```

## WorldMutation

WorldMutation 是有语义的状态变化记录。

```text
WorldMutation:
  mutation_id
  kind
  actor?
  targets
  source
  reason
  phase
  durable?
  payload
```

常见 mutation：

```text
DamageMutation:
  target
  amount
  damage_type
  source
  can_kill

MoveMutation:
  actor
  from
  to
  transition_id
  reason

InventoryMutation:
  item
  from_container
  to_container
  actor
  reason

InterestMutation:
  actor
  resource
  delta
  reason
  durable: true

ReservationMutation:
  actor
  reservation_id
  reserve / release / consume
```

WorldMutation 不一定要一开始做成完整 event sourcing，但至少应让高风险状态变化有统一记录和统一入口。

## 哪些必须走受控入口

必须走对象行为、服务或 WorldMutation：

```text
位置变化
容器 / 背包 / 装备变化
物品所有权和绑定
金钱、经验、任务关键进度
玩家利益确认点
HP / 死亡 / 复活
副本 reservation / settlement / rollback
Transition
交易
拍卖、邮件、仓库
可见世界事实，例如门状态、重要机关状态
```

可以先不走 mutation 的低风险状态：

```text
临时 AI cache
路径搜索中间结果
Presentation cadence memory
session UI focus
客户端本地滚动位置
可重算 derived view
debug-only temporary values
```

原则：

> 会影响玩家、世界事实、持久化或规则判断的状态变化必须受控；纯缓存和可重算临时值可以轻量处理。

## 与 phase 的关系

不同 phase 可以提交不同 mutation，但不能越界。

```text
Command Phase:
  主动命令引发的 Transition、Inventory、Trade、Interest mutation

Immediate Reaction Phase:
  刚发生事实引发的 Damage、Status、Trap mutation

Scheduled Event Phase:
  到期事件引发的 AutoCloseDoor、Decay、CompleteCast mutation

Active / Passive Phase:
  active system 或持续效果引发的 Damage、Duration、Recovery mutation

Settlement Phase:
  Death、DungeonSettlement、Reservation、Interest mutation

Presentation / State Flush:
  不提交 world mutation
```

Presentation / State Flush 只能读最终状态，不能补做规则逻辑。

## 面向对象不是绕开事务

面向对象不是让每个对象随便互相改状态。

不推荐：

```text
actor.location = new_place
place.occupants.append(actor)
item.container = actor.inventory
actor.money -= cost
```

推荐：

```text
TransitionService.Execute(plan)
InventoryService.PickUp(actor, item, place)
Wallet.Charge(actor, cost, reason)
DungeonService.ReleaseReservation(run)
```

对象提供行为，服务协调事务，WorldContext 维护全局一致性。

## 当前结论

```text
采用面向对象领域模型。
对象拥有自己的状态、不变量和行为。
跨对象操作通过服务或 WorldMutation 处理。
高风险状态变化必须受控。
低风险缓存和可重算临时值可以轻量处理。
Presentation 阶段不能修改 world state。
```

后续还需要讨论：

```text
Go 中对象模型如何表达：struct + methods，还是 service-heavy
Entity/Component 与面向对象行为如何结合
WorldContext 应该暴露哪些 mutation API
对象方法是否允许直接提交 PresentationEvent
事务失败如何回滚或返回 structured blocker
```
