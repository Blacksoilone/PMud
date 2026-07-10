# 时间系统设计笔记

这份笔记记录时间系统的框架层共识。它不决定具体战斗节奏、每个状态持续多久、刷新间隔是多少；这些属于后续游戏内容和玩法设计。

## 核心结论

我们采用真实 MUD 中更成熟的分层做法：**被动世界时间 + 显式调度 + 查询式冷却**。

```text
WorldClock:
  推进世界逻辑时间、mud hour、昼夜、天气、区域年龄等被动世界节律。

Scheduler:
  处理未来要发生的一次性或重复事件，例如施法完成、门自动关闭、尸体腐烂、区域 reset。

Cooldown / RateLimit:
  在命令入口查询动作是否可用，限制玩家输入、技能冷却、动作恢复和脚本刷命令。

Active Systems:
  只推进活跃集合，例如正在战斗的角色、活跃房间、活跃 NPC、活跃天气区。

Lazy Projection:
  对长期或离线变化按时间差投影，例如资源恢复、作物成长、离线训练、长期冷却。
```

原则：

> tick 推进被动世界时间；Scheduler 唤醒未来事件；Cooldown/RateLimit 限制主动命令；长期状态优先按需投影。不要把所有行为塞进一个全局 tick。

## Tick 不是命令执行

世界时间以 tick 为逻辑推进单位，但 tick 不是玩家命令的结算边界，也不是所有系统必须等待的“时钟上升沿”。

玩家动作不需要等到下一个 tick 才结算。例如：

```text
玩家移动
拿取物品
装备物品
打开容器
说话
查看物品
```

这些动作应该进入 world loop 后，按命令顺序和权威状态直接验证、执行、提交。

tick 主要服务于：

```text
世界时间推进
天气与昼夜
区域年龄
低频维护
活跃战斗轮次
活跃 NPC 行为节奏
活跃环境变化
```

原则：

> tick 是被动时间推进机制，不是主动命令限速机制。

## 第一版执行模型：顺序 phase

第一版不采用严格的 rising-edge snapshot 作为默认语义，也不为了并行而让所有被动系统只读取上一帧状态。

原因是 MUD 虽然是离散文本游戏，但玩家仍然对即时因果很敏感。如果所有被动规则都延后一帧，容易出现穿帮感。

例如：

```text
玩家进入火焰房间。
火焰系统只读上一帧，所以这一帧没有伤害。
玩家立刻离开。
结果完全没被火烧。
```

这不符合直觉。

因此第一版采用单 world loop / 单 owner 的顺序 phase 模型：

```text
1. Input Phase
   读取 session 输入，形成 ActionAttempt。

2. Command Phase
   按权威顺序执行主动命令。
   移动、拾取、交易、使用物品、开关门等在这里验证和提交。

3. Immediate Reaction Phase
   对刚发生的事实做即时环境反应。
   例如进入火焰房间、触发陷阱、接触水、进入禁区、被门夹住。

4. Scheduled Event Phase
   处理到期 ScheduledEvent。
   事件执行前重新验证目标、状态版本和当前规则。

5. Active System Phase
   推进活跃战斗、活跃 NPC、活跃房间效果、活跃天气区等。

6. Passive Effect Phase
   处理持续性、周期性、可延迟的效果。
   例如中毒、饥渴、寒冷暴露、buff duration、资源恢复。

7. Settlement Phase
   处理死亡、结算、清理、pending removal 等收尾。

8. Presentation / State Flush
   推送结构化 PresentationEvent、state patch 和必要 snapshot。
```

这个顺序本身就是规则。系统之间不能依赖隐式遍历顺序，而应依赖明确 phase。

原则：

```text
即时因果走 reaction。
持续变化走 passive phase。
玩家权益变化走 transaction。
```

### Phase contract

每个 phase 都应有明确边界。系统不能随意选择“方便的地方”修改状态。

#### Input Phase

职责：

```text
读取 session 输入
做连接级限速和队列上限检查
把输入包装成 ActionAttempt
```

允许：

```text
更新 session 输入队列
丢弃明显超限输入
生成 parser 所需的原始命令记录
```

禁止：

```text
修改 world state
移动角色
扣资源
执行命令效果
```

#### Command Phase

职责：

```text
解析 ActionAttempt
解析目标和歧义
验证主动命令
执行玩家主动操作
```

适合：

```text
移动
拿取物品
装备/卸下
打开/关闭门
使用物品
发起 TransitionRequest
交易、拾取、资源消耗等权益相关事务
```

要求：

```text
必须按权威顺序执行
同一资源竞争必须在这里解决
玩家权益变化必须事务化
```

禁止：

```text
扫描全世界
执行长时间后台逻辑
直接输出自然语言文本
```

#### Immediate Reaction Phase

职责：

```text
处理刚提交事实带来的即时环境反应
```

适合：

```text
进入火焰房间触发灼烧
踩到陷阱
进入禁区触发守卫
物品接触水后熄灭
门关闭时处理夹住的人或物
```

要求：

```text
只处理刚刚发生的局部事实
反应必须有明确来源
不能把周期性效果塞进这里
```

禁止：

```text
处理全区域 AI
处理长期环境变化
做与刚发生事实无关的状态推进
```

#### Scheduled Event Phase

职责：

```text
从 Scheduler 取出到期事件
验证目标、状态版本、generation 和当前规则
执行仍然有效的事件
```

适合：

```text
施法完成
门自动关闭
尸体腐烂
陷阱重新上膛
区域 reset 检查
长期任务到期
```

要求：

```text
每轮有执行预算
事件过期时直接丢弃
重复事件重新 schedule
```

禁止：

```text
无限补跑 missed events
在一个事件里执行大量无关工作
假设事件到期就必然成立
```

#### Active System Phase

职责：

```text
推进明确活跃的系统
```

适合：

```text
activeCombatList
activeMobList
activeRoomList
activeWeatherZoneList
```

要求：

```text
只处理 active set
每类 active system 有预算
超出预算延后
```

禁止：

```text
遍历所有 room
遍历所有 NPC
遍历所有 item
把 inactive world 唤醒成全量模拟
```

#### Passive Effect Phase

职责：

```text
处理持续性、周期性、可延迟的效果
```

适合：

```text
中毒周期伤害
饥饿、口渴、疲劳
寒冷或炎热暴露累积
buff/debuff duration
资源恢复
低频环境影响
```

要求：

```text
尽量局部、可预算
同类效果的顺序应明确
未来可演进为 EffectDelta + reducer
```

禁止：

```text
处理即时进入反应
处理拾取、交易、门票、副本结算等权益事务
执行位置变更，除非生成 TransitionRequest 交给权威流程
```

#### Settlement Phase

职责：

```text
统一处理收尾和结算
```

适合：

```text
死亡触发
DeathSession 创建
副本 run settle / rollback
清理 pending removal
释放 reservation
处理最终权益提交
```

要求：

```text
玩家利益变化必须事务化
清理顺序必须确定
同一 actor 一轮内多次死亡/结算只应处理一次
```

禁止：

```text
在结算后继续让其它 phase 修改同一批状态
把 presentation 当作结算事实来源
```

#### Presentation / State Flush

职责：

```text
把本轮产生的结构化事件和状态变化发送给客户端
```

适合：

```text
PresentationEvent
state.patch
scope snapshot
log.event
candidate / confirmation prompt
```

要求：

```text
只读最终 world state
合并 dirty fields
节流推送
按 version / seq 标记
```

禁止：

```text
修改 world state
把文本输出作为规则结果
在 flush 阶段临时补做规则逻辑
```

### Rising-edge snapshot 的位置

Rising-edge snapshot 可以作为未来优化，但不作为第一版默认语义。

适合未来采用 snapshot / delta / reducer 的系统：

```text
中毒每 tick 扣血
饥渴推进
寒冷暴露累积
天气影响
buff/debuff duration
资源恢复
普通 NPC 感知准备
```

不适合 rising-edge 并行化的系统：

```text
玩家主动命令
位置变更
拾取同一物品
交易
门票 reserve / consume
副本结算
死亡结算
打开/关闭同一扇门
多人抢资源
```

第一版可以先顺序执行 passive phase，但保持边界清晰：

```text
短期:
  active systems 顺序修改 world state。

中期:
  passive systems 输出 EffectDelta，由 reducer 合并。

未来需要时:
  部分 passive systems 并行计算 delta。
```

只有在 active entity 数量很大、tick 延迟成为真实瓶颈、需要多核利用或确定性 replay 时，才值得引入完整 snapshot / delta / reducer 模型。

## Tick 长度不是固定现实秒

tick 不一定等于一秒。它是游戏逻辑时间单位，可以根据游戏模式调整。

例如：

```text
强调即时紧张感的战斗:
  tick 可能是 0.1 秒或 0.25 秒

普通 MUD 世界:
  tick 可能接近 1 秒或更长

回合制系统:
  tick 甚至可以在回合结束时才推进
```

因此，文档和代码应避免把 tick 直接命名或理解为 real second。可以使用：

```text
WorldTick
LogicTick
GameTick
```

而不是把所有机制写死成 `server second`。

## 主动命令的速率限制

tick 不负责主动命令的防滥用。玩家动作不等 tick 才结算，但主动命令必须有独立的速率限制、队列背压和输出保护。

否则玩家可以用脚本高速发送命令，例如：

```text
nsnsnsnsnsns
```

这类输入即使每条命令都很简单，也可能造成 parser 压力、world loop 队列堆积、日志膨胀、网络输出爆炸，最终变成 DoS 攻击。

框架应支持多层保护：

```text
Connection Rate Limit:
  单个 session 每秒最多接受多少完整命令，允许短 burst，但长期限速。

Session Queue Limit:
  每个 session 未处理命令队列有最大长度。

Actor Action Gate:
  同一 actor 同时只能有有限个 blocking action；战斗硬直、施法准备、移动延迟在这里表达。

Command Class:
  look / inventory / help 这类非阻塞命令，可以和 attack / cast / move 使用不同限速。

Output Backpressure:
  如果客户端不读取输出，限制输出、丢弃低价值文本或断开连接。

Abuse Handling:
  连续超限时丢弃命令、警告、临时静默或断线。
```

原则：

> tick 处理被动时间推进；主动命令通过限速、队列上限、动作级约束和输出背压防滥用。

## Scheduler：未来事件不是全局扫描

框架不应该每个 tick 扫描全世界所有房间、物品、容器和状态。需要未来执行的行为应显式注册到 Scheduler。

不推荐：

```text
每 tick 遍历所有房间
每 tick 遍历所有物品
每 tick 递归扫描所有容器
每 tick 检查所有 tag
每 tick 调用所有 Entity.Tick()
```

推荐：

```text
点燃火把 -> 注册 ExtinguishEvent
施法开始 -> 注册 CompleteCastEvent
门打开 -> 注册 AutoCloseDoorEvent
尸体生成 -> 注册 DecayCorpseEvent
区域刷新 -> 注册 AreaResetEvent
```

原则：

> 时间系统应由显式注册的未来事件和活跃集合驱动，而不是全世界扫描驱动。

## next_wake_tick 是调度索引

如果实体有未来行为，可以在状态中记录 `next_wake_tick`，但它不是用来让引擎每 tick 扫描所有实体的字段。

更准确的模型是：

```text
Entity TemporalState:
  next_wake_tick
  wake_reason
  version

Scheduler:
  EventID
  target_entity
  wake_tick / wake_at
  priority
  seq
  target_version
  event_kind
```

不变量：

```text
如果实体声明 next_wake_tick > 0，Scheduler 中必须有对应 entry。
引擎按 Scheduler 顺序唤醒实体，不扫描所有实体寻找 next_wake_tick。
```

这样 `next_wake_tick` 是可审计的状态索引，而不是全局轮询机制。

## ScheduledEvent 与状态版本

ScheduledEvent 是“到时检查并尝试执行”的请求，不是必然事实。

例如火把被点燃时，不必每 tick 扣一次燃料；可以根据当前燃料量和燃烧强度计算预计熄灭时间，然后注册一个 `ExtinguishEvent`。

```text
点燃火把:
  BurningState.version = 17
  expected_end_tick = now + burn_duration
  schedule ExtinguishEvent(torch, expected_end_tick, state_version=17)
```

问题是：火把可能在中途被水浇灭、被补充燃料、被风吹旺、被放进特殊环境，原先的预计熄灭时间就不再可靠。

解决方式是：ScheduledEvent 应携带状态版本、事件令牌或 generation。

```text
火把被水浇灭:
  remove BurningState 或 BurningState.version++
  旧 ExtinguishEvent 仍留在队列中

旧 ExtinguishEvent 触发时:
  检查 torch 是否仍 Burning
  检查 BurningState.version 是否等于事件携带的 state_version
  如果不匹配，说明事件过期，直接跳过
```

这种模型避免了复杂的事件删除和队列内搜索，也避免了每 tick 扫描所有燃烧物。

同理适用于：

```text
冷却结束
门自动关闭
尸体腐烂
临时 buff 消失
区域 reset
陷阱重新上膛
```

原则：

> Scheduler 负责按时间唤醒；事件执行前必须重新验证目标、状态版本和当前规则。

## Active Systems：只推进活跃集合

有些系统确实需要周期推进，例如战斗、活跃 NPC、活跃房间、天气区或正在扩散的环境效果。它们可以使用 active list，而不是全局扫描。

```text
activeCombatList:
  只包含正在战斗的 actor 或 combat instance。

activeRoomList:
  只包含有玩家、近期事件或显式激活的房间。

activeMobList:
  只包含需要周期 AI 的 NPC。

activeWeatherZoneList:
  只包含当前需要推进天气或环境状态的区域。
```

active system 必须有预算：

```text
每 tick 最多处理 N 个 combat
每 tick 最多处理 N 个 NPC
每 tick 最多处理 N 个 room effect
超出预算的部分延后
```

原则：

> 周期系统只能作用于 active set，并且必须有每轮预算。inactive zone 使用 lazy projection 或下一次显式唤醒。

## Lazy Projection：长期变化按需结算

长期或离线变化不应默认注册大量长期 timer。能通过公式算出来的状态，应按需投影。

适合 lazy projection 的场景：

```text
作物成长
资源点恢复
矿脉再生
离线训练
商队长期移动
房间长期腐化
长冷却
无人区域生态变化
```

模型：

```text
state:
  last_projected_game_time
  growth_rate / recovery_rate / rule_params

on observe / activate / login:
  elapsed = now_game - last_projected_game_time
  project state forward by elapsed
  last_projected_game_time = now_game
```

这样重启、离线和无人区域都自然支持，不需要服务器一直为每个对象跑 timer。

原则：

> 长期变化优先存参数和上次投影时间；玩家观察、区域激活或系统需要时再结算到当前时间。

## Ticker 的边界

Ticker 是多个订阅者共享同一 interval 的周期调度器。它只用于确实需要周期发生的系统，不用于“检测有没有变化”。

适合：

```text
天气广播
低频区域生态刷新
全服公告节奏
周期存档
指标采集
少量共享节奏的环境系统
```

不适合：

```text
检查每个玩家是否能再次使用技能
检查每个物品是否腐烂
检查每个 NPC 是否看见玩家
检查每个房间是否需要 reset
```

这些应使用 cooldown 查询、Scheduler、事件 hook、active set 或 lazy projection。

原则：

> Ticker 是低频共享节奏工具，不是变化检测工具。

## 崩溃恢复与补算

服务器暂停或重启后，不应该无上限补跑过去所有 tick。

策略：

```text
世界时间可以跳到当前。
过期的一次性事件按 due time 进入恢复队列。
重复事件默认 coalesce，不逐帧补跑。
长期状态通过 lazy projection 结算。
战斗、AI 这类敏感系统不补跑大量历史帧，只恢复到安全状态或重新调度。
```

每轮恢复也必须有预算，避免启动时被大量过期事件拖垮。

原则：

> 恢复时结算“现在应该是什么状态”，不要机械补跑所有错过的 tick。

## 后台任务与管理能力

MOO/MUSH 的经验说明：任何可延迟、可排队、可由脚本触发的任务，都必须有管理边界。

ScheduledEvent / background job 应至少包含：

```text
job id
owner id
target id
event kind
due time
priority
sequence
state version / generation
created at
schema version
```

框架应准备：

```text
列出某 owner / target 的 pending jobs
取消单个 job
取消某 owner / target 的 jobs
暂停后台 dequeue
每轮最大事件数
同 owner 同 tick 最大事件数
单事件最大执行耗时记录
失败事件日志或 dead-letter
```

原则：

> 只要任务能排队，就必须能限额、观察和取消。

## 框架需要准备什么

框架层应准备：

```text
WorldClock / LogicClock
Tick 推进接口
Scheduler / ScheduledEvent 队列
ScheduledEvent state version / event token
ActiveSystem 注册与预算
Ticker 订阅机制
Cooldown / RateLimit 查询接口
LazyProjection 状态接口
持久化 scheduled event 的能力
崩溃恢复后的时间恢复策略
主动命令限速
session 命令队列上限
输出背压策略
动作级冷却/成本接口
后台任务管理接口
```

框架层不应该现在决定：

```text
tick 具体多长
战斗是不是即时制
是否回合制
燃烧多久结算一次
NPC 每几 tick 行动
区域多久 reset
具体 cooldown 数值
具体资源恢复速度
```

这些属于玩法和内容设计。

## 何时使用什么

```text
技能 10 秒后可再用:
  Cooldown expires_at
  不注册未来 callback

施法 3 秒后完成:
  ScheduledEvent continuation
  不阻塞 goroutine

天气每 5 分钟变化:
  Ticker 或 ScheduledEvent
  不每 tick 扫描所有房间

NPC 只有玩家进入房间才反应:
  OnEnterRoom hook + active set
  不让所有 NPC 每秒查房间

区域 30 分钟后刷新:
  ScheduledEvent + occupancy check
  不每 tick 扫所有 zone

自动保存:
  低频 scheduled event
  不挂在命令执行路径

战斗 round:
  activeCombatList 或 per-combat scheduled event
  不扫所有 actor

作物成长 / 矿脉恢复:
  LazyProjection
  不为每个作物或矿脉长期挂 timer
```

## 当前结论

当前时间系统框架结论是：

```text
世界时间以 tick 为逻辑推进单位。
tick 不一定等于现实秒。
tick 只推进被动世界时间和活跃系统。
玩家动作不等待 tick。
玩家主动命令需要独立限速、动作 gate 和输出背压。
未来行为进入 Scheduler，而不是全世界扫描。
next_wake_tick 是 Scheduler 索引，不是轮询字段。
持续状态使用 state version / event token 让旧事件安全失效。
周期系统只处理 active set，并且必须有预算。
长期变化优先 lazy projection。
Ticker 只用于低频共享节奏，不用于变化检测。
崩溃恢复不无限补跑 missed ticks。
后台任务必须可限额、可观察、可取消。
```

后续还需要讨论：

```text
WorldClock 的具体 tick 粒度
Scheduler 第一版使用 min-heap 还是数据库索引
哪些事件需要跨重启持久化
combat round 是 active list 还是 per-combat scheduled event
AreaReset 与 live state 的关系
哪些系统默认 lazy projection
```
