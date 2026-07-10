# MUD 服务端架构笔记

这份笔记总结了对经典和现代 MUD 服务端的调研，重点是：如果要用 Go 手写一个小型 MUD 服务端，应该吸收哪些成熟设计，避免哪些框架化复杂度，以及怎样把“框架”“内容”和“运行中的世界状态”分开。

目标不是复制现成框架，而是理解 MUD 背后的稳定架构思想，并为后续手写实现建立清晰边界。

## 核心结论

MUD 服务端不应该把玩家利益当成一份临时的进程内存。服务器重启或崩溃后，临时 live world 可以重建或 reset，但玩家已经确认获得、消耗或被承诺的长期利益不应该无解释回退。

这里至少有三层概念，必须分清：

1. **引擎或框架**：Go 写的运行时，负责接受连接、解析命令、调度事件、加载内容、保存状态、恢复状态、执行规则。
2. **内容定义**：设计出来的世界，包括房间模板、出口、物品定义、NPC 定义、技能数值、刷新规则、对话、任务规则。
3. **运行中的世界状态**：这个世界实际正在发生什么，包括玩家在哪里、哪些门被打开、哪些物品被移动、哪些怪物死亡、哪些副本实例正在运行。
4. **玩家利益状态**：已经确认属于玩家或已经从玩家处确认消耗的长期事实，包括物品、金钱、经验、任务关键进度、仓库、邮件、交易、门票 reservation、结算后的副本奖励。

经典 MUD 代码库有时会把这些层混在一起。有些老 MUD 允许 builder 直接在持久世界数据库里编辑内容；Diku/CircleMUD 一类系统则经常从静态区域文件加载世界，然后按 reset 规则刷新区域。

对于一个新的 Go 服务端，更清晰的模型应该是：

```text
Go 引擎
  加载带版本的内容定义
  恢复已持久化的玩家利益状态和必要的运行状态
  运行世界模拟
  周期性 checkpoint 关键状态
  把重要事件追加到持久存储
```

内容文件本身不等于持久化。如果服务器启动时只是重新读取房间、物品、NPC 定义，那么它恢复的是“设计时的世界”。这对门、尸体、普通地面掉落、NPC 临时位置、副本内部状态等临时 live world 可以接受；但对玩家已经确认的利益不可以接受。

本项目的持久化目标不是“完整冻结 live world”，而是“保证玩家利益边界不回退”。临时世界可以 reset；已提交的玩家权益不能因为重启消失。

## 不同 MUD 家族的启发

### Diku / CircleMUD 风格

Diku 派生服务端通常使用明确的 C 结构体表示房间、物品、怪物、角色、descriptor、zone。网络循环从 descriptor 读输入，把输入交给命令解释器，命令 handler 再修改世界状态。

内容通常存在 area/world 文件里，例如房间、mob、object、shop、zone reset 命令。

最值得借鉴的是它的简单运行形状：

```text
descriptor 输入
  -> command interpreter
  -> command handler
  -> room/object/mobile/character 状态
  -> output queue
```

它的 reset 系统也很重要。很多 Diku-like 游戏的世界本来就是半重置的：zone 文件描述区域如何重新填充，比如在某个房间刷一个 mobile，把某个 object 放进容器，重置某扇门。

但这不是崩溃恢复。reset 是游戏机制，持久化是运行保障。

可以借鉴：

- 命令表
- descriptor/session 分离
- room/mobile/object 模型
- zone 或 area 内容文件
- 把 reset 作为显式游戏机制

早期应避免：

- 到处都是全局可变状态
- 巨大的命令 handler 直接了解所有子系统
- 误以为静态 area 文件就是持久化

### LambdaMOO / MOO 风格

LambdaMOO 把世界看成一个持久对象数据库。对象有属性和 verb，运行中的世界不是每次启动都从静态文件重建。

这对持久化和游戏内建造很有启发。

可以借鉴：

- 世界可以是持久对象图
- 运行时服务端和持久世界数据是两件事
- 用户或 builder 可以在不重新编译服务器的情况下修改世界

早期应避免：

- 自己发明一门编程语言
- 在基础游戏跑起来前就实现通用对象数据库

### LPMud / FluffOS 风格

LPMud 把系统分成 driver 和 mudlib。driver 处理底层运行时，mudlib 用 LPC 定义游戏行为。即使不复制 LPC，`heart_beat`、`call_out`、`reset`、房间脚本、对象脚本这些概念也很有用。

可以借鉴：

- 引擎和内容分离
- 定时调用和 heartbeat 类事件
- reset hook 作为内容行为
- 游戏逻辑不写进底层网络引擎

早期应避免：

- 在稳定的引擎 API 出现之前，就让每个房间都变成脚本文件

### MUSH / PennMUSH / TinyMUX 风格

MUSH 系统重度依赖持久对象数据库、attribute、softcode、权限、命令队列和用户创建内容。它们很适合作为持久化、权限、配额、builder 工具的参考。

以后可以借鉴：

- dbref 或稳定实体 ID
- 权限检查
- 持久 attribute
- 命令配额和队列限制

早期应避免：

- softcode 复杂度
- 在世界模型还不清楚时开放太多可编程能力

### 现代框架

Evennia、Ranvier、CoffeeMUD、AresMUSH 等现代框架最值得学习的是边界，而不是完整复杂度。

Evennia 的概念分层很有价值：protocol/session、account、character、object、command、cmdset、script、persistent entity。Ranvier 对内容 bundle、YAML 房间、脚本化行为的组织很有参考意义。CoffeeMUD 展示了一个全功能 MUD 引擎会膨胀到什么程度，它更像一个警示。

可以借鉴：

- session 不是 player
- account 不是 character
- protocol 不是游戏逻辑
- command parsing 不是 world simulation
- 内容包需要结构
- 管理工具和 builder 工具比想象中更早重要

早期应避免：

- 框架级插件系统
- 在 Go 里复制动态 typeclass 体系
- 在核心 MUD loop 之前先做 Web UI、账号系统和复杂权限

## 玩家身份、Session 与角色

网络连接、账号、角色和世界中的玩家实体必须分开。

```text
Session:
  一次网络连接或客户端连接
  负责输入、输出、认证状态、断线和重连
  不等于玩家角色

Account:
  长期身份
  拥有一个或多个 Character
  第一版可以很薄，但边界要保留

Character:
  玩家长期角色档案
  保存名字、基础属性、经验、装备、背包、任务进度等玩家利益

PlayerEntity:
  Character 在 world loop 中的运行时实体
  有当前位置、当前状态、战斗状态、临时效果和 observation scope
```

推荐连接链路：

```text
Client connection
  -> Session
  -> Account
  -> Character
  -> PlayerEntity in World
```

第一版采用混合模式：架构上保留 `Session / Account / Character / PlayerEntity` 的完整边界，但开发期允许匿名或 dev session 自动创建临时账号和临时角色。

```text
正式结构:
  Session != Account
  Account != Character
  Character != PlayerEntity

开发便利:
  dev session 可以自动创建 temporary Account
  temporary Account 可以自动创建 temporary Character
  temporary Character 可以直接进入默认起始地点
```

这样早期开发不需要完整账号 UI，也不会把连接和角色绑死。之后要支持重连、多角色、管理员观察模式、同账号互斥登录、游客模式或正式账号迁移时，不需要推翻世界模型。

开发期 anonymous / dev 身份只是开发者测试工具，不是普通玩家的游客账号。

```text
temporary Account / Character:
  developer-only
  test-only
  not player-facing guest mode
  cannot upgrade to formal Account
  cannot bind to email or external identity later
  can be wiped freely by dev/test reset
```

原因：测试角色可能拥有调试权限、跳过正常经济约束、经历不稳定内容或被开发工具修改。如果允许这些角色升级成正式账号，会污染正式玩家经济和进度边界。将来即使引入邮箱验证或正式账号体系，测试身份也不迁移。

### 一个账号多个角色

一个 `Account` 应允许拥有多个 `Character`。

原因：MUD / MMO 经常包含永久性或半永久性的选择，例如种族、职业、阵营、信仰、出生地、死亡路线、不可逆任务分支、声望关系等。想体验不同玩法的玩家不应该被迫创建新的账号、邮箱或认证身份。

```text
Account:
  owns many Characters

Character:
  belongs to one Account
  has independent progression, inventory, equipment, quests, location
```

第一版 UI 可以先做非常简单的角色选择：

```text
dev / anonymous mode:
  自动选择或自动创建默认 Character

normal mode:
  Account 登录后列出 Characters
  玩家选择一个 Character 进入世界
  如果没有 Character，则创建一个
```

限制多角色的地方应该来自游戏规则，而不是账号系统。例如同账号角色之间是否能互相交易、是否共享仓库、是否共享账号成就、是否允许同账号多开，这些是后续 gameplay / economy policy，不应该通过“一个账号只能一个角色”来硬塞。

原则：

> Account 是身份容器，不是单个角色。一个账号允许多个角色，让玩家能体验永久选择和不同玩法，而不需要创建多个邮箱或认证身份。

原则：

> Session 是连接，Account 是身份，Character 是长期玩家档案，PlayerEntity 是世界中的运行时角色。第一版可以自动创建 dev 身份，但不能把这些概念合并成一个对象。

### 断线与重连

断线不等于角色立刻离开世界。第一版默认采用短暂保留模型：

```text
Session disconnected:
  PlayerEntity 标记为 disconnected
  角色短暂留在 world loop 中
  开始 reconnect grace window

Session reconnected within window:
  新 Session 重新绑定同一个 Account / Character
  恢复控制原 PlayerEntity

grace window expired:
  根据当前位置、战斗状态、当前 run、特殊环境、特殊状态和内容规则结算
```

普通安全场景下，超时后可以执行 logout：

```text
not in combat + safe logout allowed:
  保存 Character 状态
  移除 PlayerEntity 或标记 offline
```

危险或特殊场景下，超时不能被当作免费逃生：

```text
in combat / dungeon / special environment / special status:
  按场景策略处理
  可能 fallback escape
  可能 fail dungeon run
  可能判定为死亡
```

判定原则不是“区域危险不危险”，而是：玩家离线后是否会带来显著的可利用收益。

```text
offline creates significant exploitable benefit:
  断线超时应视为失败或死亡

offline does not create significant exploitable benefit:
  断线超时不应死亡，按普通 logout / fallback 处理
```

“长时间未连接判死亡”不是全局默认，也不能只因为角色位于高危区域就触发。它只适合用于离线会明显规避代价或获得收益的场景，例如战斗中断线、副本中断线、特殊环境、特殊状态、特殊仪式、倒计时机关等。

高危区域本身不构成显著收益，原因是角色下次上线时仍然在原区域，仍然要面对原本的风险。断线只是延后风险，不是取消风险。

```text
玩家在 D4 野外断线:
  下次上线仍在 D4 野外
  风险被延后，但没有被消除
  不应仅因断线超时死亡

玩家战斗中断线:
  可能规避正在发生的伤害、追击、资源消耗或死亡结算
  可以视为显著可利用收益

玩家副本中断线:
  可能规避限时、团灭、机关失败、消耗或结算失败
  可以视为显著可利用收益
```

如果断线后找强大玩家来清理附近怪物，这通常也不是断线带来的显著收益，因为这些玩家本来就可以在一开始组队或支援。它没有创造原本无法获取的收益，只是改变了协作时机。

原则：

> 断线先给短暂重连窗口；窗口后按离线可利用性结算。普通断线不自动死亡，高危区域本身也不构成断线死亡条件；如果离线会显著获利或规避代价，则对应场景应显式声明长时间未连接等同失败或死亡。

### 角色绑定与踢旧连接

同一个 `Character` 同一时间只能有一个 controlling `Session`。

```text
same account + same character + old session disconnected:
  新 Session 接管原 PlayerEntity
  old Session 作废

same account + same character + old session still connected:
  允许新 Session 踢掉旧连接并接管
  old Session 进入 detached / closed
  old Session 之后不能再提交命令

different account + same character:
  拒绝绑定
  除非是明确的 admin / debug / observer 权限
```

这避免同一个世界实体被多个客户端同时控制，也解决客户端卡死、网络假连接、设备切换时玩家被旧 Session 锁在外面的问题。

推荐把这件事集中到 `CharacterBindingService`：

```text
Client connects
  -> create Session

Session authenticates
  -> identify Account

Account selects Character
  -> CharacterBindingService.Bind(session, account, character)

Bind result:
  bind_existing_disconnected_entity
  kick_old_session_and_bind
  create_or_load_player_entity
  reject_character_controlled_by_other_account
```

原则：

> 一个 Character 只有一个控制连接。同账号新连接可以踢掉旧连接；不同账号不能控制同一角色，除非是显式管理或观察权限。

## 推荐的 Go 运行时形状

对于小型 Go MUD，最清晰的模型不是高性能事件循环网络库，而是先用标准库：

```text
net.Listener
  -> accept loop
    -> 每个连接一个 session
      -> read pump goroutine
      -> write pump goroutine

所有解析后的命令
  -> 有界 command channel
    -> 单个 world loop goroutine
      -> 修改权威世界状态
      -> 产生输出消息
        -> session write queue
```

最重要的规则是：session goroutine 不直接修改世界状态。它们只负责把网络输入翻译成命令，再把命令发给 world loop。world loop 拥有可变的模拟状态。

这个设计的好处是：

- 数据竞争更少
- 推理更容易
- 命令顺序更确定
- tick 和 timer 有自然归属
- persistence checkpoint 更容易做
- 以后想做 replay 或 event log 更容易

如果将来真的需要扩展，可以按 area 或 zone 分片。对于一二十个玩家，单个 world loop 完全足够。

## 数据驱动世界和持久运行状态

引擎应该从数据加载内容定义，但也必须持久化运行状态。这是两种不同的存储问题。

### 内容定义

内容定义回答的是：设计出来的世界包含什么？

例子：

- 房间模板：`town.square`
- 房间标题和描述
- 从一个房间模板到另一个房间模板的出口
- 物品定义：`iron_sword`
- NPC 定义：`rat`
- 刷新规则：下水道房间可以刷老鼠
- 技能公式
- 商店库存模板
- 任务定义

这些内容适合放在可编辑文件里，例如：

```text
content/
  areas/
    town/
      rooms.toml
      mobs.toml
      items.toml
      resets.toml
```

它们可以热加载，因为它们只是定义，不一定是当前世界的真实状态。

### 运行中的世界状态与玩家利益状态

运行中的世界状态回答的是：这个实际世界现在正在发生什么。

例子：

- 玩家 A 在 `town.square`
- 物品实例 `item_8271` 掉在 `sewer.entrance`
- 门 `town.gate:north` 已经解锁
- NPC 实例 `mob_441` 已死亡，并将在时间 T 复活
- 箱子 `chest_12` 已经被掏空
- 副本实例 `run_91` 正在进行

这些状态不一定都要跨重启保存。普通地面掉落、门开关、尸体、战斗中间态、副本内部怪物血量都可以是 volatile live state。

玩家利益状态回答的是：什么已经确认属于玩家，或已经确认从玩家那里消耗。

例子：

- 玩家 B 背包里已经有 `item_8271`
- 玩家获得了经验、金钱或任务关键 flag
- 玩家仓库、邮件、拍卖、交易中的资产
- 玩家副本门票被 reserve
- 副本离开结算后发放的奖励

这一层必须持久化。服务器重启可以丢临时世界状态，但不能让玩家已经确认的利益无解释回退。

### 派生状态

有些状态不需要持久化，因为可以从内容定义推导出来：

- 房间标题
- 物品基础伤害
- NPC 最大生命值
- 房间默认出口

应该只持久化真正需要跨重启存在的 delta。除非你有意做版本化快照，否则不要把每个静态定义完整复制一份到存储里，也不要为了保存玩家利益而强行保存整个 live world。

## 实用的持久化模型

第一个正式版本可以朝三层存储方向设计：

```text
content files
  定义设计中的世界

database tables 或 durable JSON snapshots
  保存玩家利益状态和少量必要运行状态

append-only event log
  记录重要权益变更，用于恢复和调试
```

第一天不需要三层全做完。这个方向要避免的不是“临时世界 reset”，而是“服务器崩溃后玩家已确认利益回退”。

### 最小可用持久化

先从玩家持久化和确认点开始：

- 保存玩家位置、背包、属性和 flag
- 保存玩家物品、金钱、经验、任务关键进度
- 保存仓库、邮件、交易、拍卖这类长期权益
- 保存副本门票 reservation 或其它未完成承诺
- 后续可加入周期性 checkpoint；第一版先依赖确认点同步保存

第一版使用 SQLite。SQLite 提供事务和崩溃安全，又不需要额外数据库服务，适合先恢复手写代码能力并跑通玩家位置、物品实例和 relation 持久化。

SQLite 只是第一版持久化后端，不是世界模型的一部分。代码应通过 `Store` / `PersistStore` 边界访问长期事实，避免 world、command、session 直接依赖 SQL 或具体数据库。

```text
World / Command:
  只表达需要保存的玩家长期事实
  不拼 SQL
  不依赖 SQLite 连接

PersistStore:
  SaveCharacterPosition
  SaveItemInstance
  SaveItemRelation
  LoadCharacterState

SQLite backend:
  第一版实现细节

PostgreSQL backend:
  未来可替换实现
```

原则：

> SQLite is the first backend, not the persistence model. Persist player facts through a Store boundary so the database can be replaced later.

### 更好的持久化

world loop 出现之后，可以把每个有意义的玩家利益变更看成确认事件：

```text
PlayerMoved
ItemPickedUp
ItemDropped
DoorUnlocked
MobKilled
MobRespawnScheduled
QuestFlagSet
TicketReserved
TicketReservationReleased
DungeonRunSettled
```

然后可以：

- 把事件应用到内存状态
- 对玩家利益事件追加到持久存储
- 周期性写 snapshot
- 恢复时加载最新 snapshot，再 replay 之后的事件

这不意味着马上要做完整 event sourcing。只是说，如果命令从一开始就区分“临时 live state 变化”和“玩家利益确认事件”，后续持久化会容易很多。

### 确认点

玩家利益只在确认点之后成立。确认点之前的中间态可以丢；确认点之后的事实必须持久化。

开放世界中，推荐边界是：

```text
地上掉落:
  临时 live state，重启可以丢。

玩家拾取成功:
  形成确认点，物品进入背包前必须事务化持久化。
```

副本中，推荐边界是：

```text
副本内获得物:
  run-local pending reward，不是玩家长期资产。

玩家离开副本并结算:
  形成确认点，清理副本内临时物品，提交奖励到玩家持久状态。
```

原则：玩家看到“已经获得/已经消耗/已经完成”的结果前，相关持久化必须已经成功；否则只能显示 pending、暂存或副本内状态。

第一版采用同步保存确认：会改变玩家长期事实的命令，必须先通过 `PersistStore` 保存，再展示成功。保存失败则命令失败，不允许先告诉玩家成功再异步保存。

```text
get / drop / go / quit:
  如果改变角色位置、玩家物品实例或玩家物品 relation
  必须在 success PresentationEvent 前保存

look / inventory / say:
  只读或聊天
  不写持久化
```

第一版不做 event sourcing、异步 outbox 或后台延迟提交。这样吞吐量不是最高，但边界清楚，适合先恢复手写实现能力。

第一版也不做通用 checkpoint 或后台周期性保存全部玩家状态。可靠性来自确认点同步保存。graceful shutdown 只负责停止接受新连接、drain 已入队命令、关闭 session 和 flush/close 数据库资源；它不是“补保存未确认状态”的机制。

后续出现长时间 buff、战斗中状态、副本 run、定时任务、离线恢复窗口或 builder 草稿后，再讨论 checkpoint。

### 副本事务模型

副本可以看作一个可回滚的临时事务。

```text
进入副本:
  检查门票或进入资格
  reserve 门票，不真正 consume
  创建 volatile DungeonRun live state

副本中:
  怪物状态、机关状态、副本内特殊物品、pending reward 都是 run-local state
  这些状态默认不跨重启持久化

正常完成并离开:
  consume reserved ticket
  清理 run-local-only items
  提交 pending rewards
  标记 run settled

失败、放弃或服务器重启:
  释放 reserved ticket
  丢弃 pending rewards
  清理或丢弃 run-local state
```

也就是说：

```text
DungeonRun = temporary transaction
ticket reservation = precondition lock
pending reward = uncommitted changes
settlement = commit
restart / abandon = rollback
```

这样副本可以在重启后直接 reset，而不会白白损害玩家门票；同时也不需要持久化副本内所有怪物、房间、临时物品和计时器。

副本是隔离 run，不是普通开放世界区域的一部分。不同玩家进入同一个副本入口时，不一定会进入同一个 run；默认可以各自进入独立实例。只有组队或明确匹配到同一 run 的玩家，才会在副本内互相可见。

```text
单人进入:
  创建或进入自己的 DungeonRun

组队进入:
  队伍成员进入同一个 DungeonRun

非同队玩家:
  即使入口相同，也可以进入不同 DungeonRun
```

副本内掉落、临时物品、机关状态和怪物状态只属于该 run。离开、失败、rollback 或 run 结束时，副本内部状态直接清理，不进入开放世界地面清洁系统。

副本不使用开放世界房间危险度 D0-D4。副本难度是 `DungeonRun` 级别的规则，用来描述整个副本的挑战强度、推荐人数、奖励倍率和失败代价；副本房间内部只保留 encounter tuning，例如怪物强度、机关压力、boss 阶段或检查点角色。

```text
DungeonDifficulty:
  run-level challenge and reward profile

DungeonFailurePolicy:
  death / wipe / abandon / timeout settlement rules

DungeonRoomTuning:
  encounter_level / trap_intensity / checkpoint_role
```

第一版副本默认按单人或非强组队模型设计。MUD 本身不太适合一开始就做强多人副本；多人副本涉及队伍成员掉线、部分结算、队长权限、奖励归属和副本继续条件，后续再单独讨论。

副本中的主动退出和网络掉线要分开：

```text
主动 leave / quit:
  视为放弃副本
  rollback run
  释放 reserved ticket
  不提交 pending rewards

网络掉线:
  保留短暂 reconnect window
  玩家重连后可回到同一个 run
  超时未回则 rollback run

服务器重启:
  不保存 DungeonRun live state
  rollback run
  释放 reserved ticket
  不提交 pending rewards
```

这样可以保护普通网络波动，同时避免为了副本掉线恢复而持久化完整副本状态。

## 安全出口与主动脱困

很多系统都需要把玩家强制迁移到安全地点，例如副本 rollback、死亡、非法位置恢复、房间被热加载删除、特殊状态清理。不要让每个系统各自硬编码目的地，而应该通过统一的 `FallbackPlaceResolver` 选择安全地点。

推荐优先级：

```text
1. 当前玩法实例指定的 safe_exit_place
   例如副本、竞技场、特殊挑战

2. 当前房间或区域指定的 fallback_place
   例如本地区安全区、驿站、城镇入口

3. 玩家绑定点
   例如 home、门派、复活点

4. 全局默认安全点
   例如新手村广场

5. 如果以上都不存在，启动失败或管理员介入
```

典型用途：

```text
副本 rollback:
  清理 run-local state
  release ticket reservation
  移动到副本 safe_exit_place

玩家死亡:
  清理 combat / temporary state
  计算死亡惩罚
  移动到本区域 safe_respawn_place

内容热加载删除房间:
  检测当前位置非法
  移动到区域 fallback_place
```

框架也应提供玩家主动脱困命令，例如 `unstuck` 或 `escape`。它不是回城功能，而是让玩家摆脱 bug、死锁或无法继续操作的状态。

主动脱困不应要求玩家在危险或死锁状态中等待。例如有些卡住场景是“玩家无法攻击，但敌人伤害低于玩家回血”，如果要求传送前等待两分钟，反而无法解决问题。

因此推荐模型是：

```text
玩家输入 unstuck / escape:
  检查当前玩法是否允许主动脱困
  清理不能带走的临时收益状态
  立即通过 FallbackPlaceResolver 移动到安全地点
  给予 post-escape restriction

post-escape restriction:
  一段时间内不能移动，或不能离开安全区
  一段时间内不能进入战斗、开启副本、交易或拾取关键物品
  命令本身进入较长 cooldown
```

防滥用重点不是“传送前站定”，而是“传送后不能立刻把它当逃跑、回城或收益转移手段”。限制可以类似：

```text
escape 后 2 分钟不能移动
escape 后 2 分钟不能攻击或进入危险区域
escape 后清理跟随者、押送目标、副本临时物品、未结算奖励
escape 命令有较长 cooldown
某些玩法可禁止 escape，或只允许回滚到该玩法 safe_exit_place
```

原则：

> FallbackPlaceResolver 负责“去哪儿安全”；unstuck/escape 负责“玩家主动脱困”。主动脱困应先解决卡死，再通过传送后的限制和状态清理防滥用。

## 内容热加载不等于世界重置

热加载不应该表示“丢掉当前 live world，然后重新建一个”。它应该表示：

```text
加载新的内容定义
校验它们
构建 ContentSnapshot vN+1
和当前 live world 对比
应用安全的定义变化
保留 live mutable state
```

例子：如果房间 `town.square` 的描述改了，那么当前站在这个房间的玩家下次 `look` 时应该看到新描述。他们的位置不会重置。

例子：如果物品定义 `iron_sword` 的基础伤害从 5 改成 6，已经存在的铁剑实例默认通过 `definition_id` 读取 active `ContentSnapshot` 的新定义。但它自己的耐久、绑定、词条、位置和玩家权益状态不会被模板覆盖。

例子：如果内容里删除了一个房间，而当前有玩家在里面，第一版 reload 应该失败，除非新 snapshot 提供显式迁移或 fallback policy。静默删除很危险。

第一版热加载关系：

```text
LiveEntity:
  entity_id
  definition_id
  mutable_state
  player_interest_state

active ContentSnapshot:
当前定义、展示、TagClass、script binding、provider lists
```

热加载改变定义，不重置实例事实：

```text
内容描述、展示、脚本绑定、模板参数:
  可以随 active ContentSnapshot 改变

玩家位置、背包、物品实例、耐久、绑定、词条、当前状态:
  保留在 LiveState
  不被模板覆盖
```

第一版不做自动 live migration。如果新 snapshot 与当前 live world 引用不兼容：

```text
definition_id 不存在
玩家所在 place 被删除且没有迁移规则
required component schema 不兼容
script binding / resolver / provider 缺失
mutable state key 类型不兼容
```

则：

```text
拒绝新 ContentSnapshot
保留旧 snapshot 继续服务
报告内容校验错误
不做部分热加载
```

推荐的 reload 流程：

1. 监听目录，而不是只监听单个文件。
2. 对文件变化做 debounce。
3. 解析所有受影响内容。
4. 校验引用：出口、物品 ID、mob ID、reset 目标。
5. 构建完整不可变的 `ContentSnapshot`。
6. 和当前 live world 对比，发现危险变化。
7. 校验通过后，原子替换 snapshot。
8. 校验失败时，保留旧 snapshot，并报告错误。

这样即使内容编辑出错，服务器也能继续运行。

## Reset 是游戏机制，不是崩溃恢复

MUD 经常有 reset，但这不意味着整个世界应该在进程重启时 reset。

至少有几种 reset 要分开：

### 区域 reset

区域 reset 按设计规则重新填充内容：

```text
如果下水道里的老鼠少于 3 只：
  刷新老鼠

如果箱子为空且 reset 时间已到：
  重新填充箱子
```

这是游戏规则。

### 副本 reset

一个副本地下城可能在所有玩家离开后 reset，也可能按计时器 reset。这也是游戏规则。

### 服务器重启恢复

服务器重启恢复，是把玩家利益状态恢复到最近的确认点，并把必要运行状态恢复到安全状态。这是运行保障。

不要混淆它们。服务器崩溃可以让临时 live world 回到内容定义或 reset 后状态，但不能让已确认的玩家利益无解释回退。

## 稳定 ID 很重要

如果内容和 live state 是分离的，每个需要持久化的东西都需要稳定身份。

例子：

```text
room template id: town.square
exit id: town.square:north
item definition id: weapon.iron_sword
item instance id: item_0000008271
mob definition id: mob.sewer_rat
mob instance id: mob_0000004410
player id: player_0000000007
```

定义和实例是两件事。`weapon.iron_sword` 是一种物品，`item_0000008271` 是世界里实际存在的一把剑。

这个区分能避免很多持久化 bug。

## 建议的第一版架构

可以先按这些概念划分。早期不要过度纠结包名，边界比目录名更重要。

```text
server
  接受 TCP 连接
  拥有 session
  处理 shutdown

session
  读取行输入
  写出输出
  跟踪连接状态
  不直接修改 world

command
  解析输入行
  把命令名映射到 handler
  产生 world command

world
  拥有 live state
  运行 world loop
  应用 command 和 event
  产生输出消息

content
  加载 room/item/mob 定义
  校验引用
  构建不可变 snapshot

persist
  保存和加载 player-interest state
  保存确认事件和必要 reservation
  以后追加 event log
```

第一个可玩版本选择“极小可玩 MUD”，但必须非常小。每个可实现维度只做 2-3 个样例，用来验证抽象、形成对比并排除单例特例；不要为了内容量扩展。

```text
TCP 客户端可以连接
玩家出现在 2-3 个房间组成的小区域里
look 显示房间、出口、玩家和少量可见物
say 广播给同房间玩家
go 在 2-3 个出口/方向之间移动
get / drop / inventory 操作 2-3 个手工物品
quit 断开连接
服务器可以保存玩家位置
重启后恢复玩家和已确认物品利益
```

这个版本可以验证：

```text
Session / Character / PlayerEntity 绑定
World loop 命令执行
Entity relation：房间、出口、背包、地面物
Observation / Presentation 的最小输出
拾取成功后的玩家权益确认点
drop 后的地面物状态
玩家位置和背包持久化
断线重连后的角色恢复
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
泛用脚本系统
大型内容区域
```

## 早期应该避免什么

在核心 loop 和持久化模型清楚之前，先避免：

- 到处嵌入 Lua，把每个房间和物品都变成任意脚本对象
- 自制脚本语言
- Go plugin 热加载
- ECS 框架
- 高性能事件循环网络库
- 完整 telnet 协议支持
- Web 管理后台
- 在匿名 session 跑通之前就做 account/character 分离
- 巨型对象数据库
- MUSH 风格 softcode
- 在移动、持久化、内容加载之前先做战斗系统

这些功能本身不是坏东西，只是不适合作为恢复手写开发能力的第一批功能。

脚本语言的价值需要重新表述：它的核心不是比声明式数据更“强”，而是让内容行为可以免 Go 编译地迭代。Go 没有适合本项目日常内容开发的动态链接路径，不能要求内容作者为了改一个房间机关、特殊物品反应或 NPC 行为就完整重新编译服务端。

因此早期应避免的是“无边界脚本化”，不是永远拒绝脚本：

```text
应该保持声明式:
  tags
  components
  relations
  display metadata
  common requirements

可以作为热加载脚本出口:
  special trigger reaction
  room mechanism
  NPC dialogue step
  quest step
  custom resolver
```

脚本必须通过稳定 Go API、ContentSnapshot 和 sandbox 进入系统，不能绕过 world transaction、持久化和权限边界。

内容组织上，声明式内容是主入口，Lua 脚本是被内容引用的行为模块。区域专属行为放在 `areas/<area>/scripts/` 附近，多个区域复用的行为放在 `shared/scripts/`。脚本不反向定义实体，不声明 tag，不成为隐藏的内容来源。

## 推荐的学习实现路径

这条路径应该迫使你亲手写关键部件，同时避免一开始淹没在不必要的系统里。

### 里程碑 1：网络 shell

- 接受多个 TCP 客户端
- 读取行输入
- 写出响应
- 支持 `quit`
- 干净地断开慢客户端或异常客户端

### 里程碑 2：session 和 world loop

- 每个连接变成一个 session
- 输入被转成 command
- world loop 拥有 live state
- `look` 和 `say` 在一个房间里可用

### 里程碑 3：房间图

- 房间有稳定 ID
- 出口连接房间
- `go <direction>` 移动玩家
- `look` 显示当前房间和出口

### 里程碑 4：内容定义

- 从数据文件加载房间和出口
- 启动时校验引用
- 如果想保持小步前进，这一阶段可以只做玩家持久化

### 里程碑 5：持久化

- 保存玩家位置
- 保存玩家背包、装备、金钱、经验和关键 flag
- 保存拾取成功后的物品变化
- 保存副本门票 reservation 和未完成承诺
- 重启后恢复玩家已确认利益，临时 live state 可重建

### 里程碑 6：热加载

- 重新加载内容定义，但不替换 live state
- 拒绝非法内容 reload
- 明确删除和迁移规则

### 里程碑 7：reset 规则

- 区域 reset 按设计重新填充内容
- reset 操作 live state
- reset 不是崩溃恢复

### 里程碑 8：行为规则

- 先定义声明式 trigger/action 和稳定行为 API
- 需要频繁内容迭代的行为，允许接入受限热加载脚本
- Lua / Starlark / Tengo 的选择后置，但架构从一开始为脚本绑定留出口

## 当前建议

这个项目的第一版设计，建议是：

```text
Go 引擎
标准库 TCP
session read/write pump
单个 world loop
数据驱动的内容定义
SQLite 或 durable snapshot 保存玩家利益状态
显式 reset 规则
通过不可变 ContentSnapshot 做热加载
第一版不做泛用脚本系统，但保留受控脚本绑定边界
```

最重要的架构承诺应该是：

> 设计中的世界可以从内容文件重新加载，临时 live world 可以在重启后重建；但玩家已确认的利益必须从持久化状态恢复。重启服务器可以重置场景，不能无解释回退玩家权益。

这个承诺应该从一开始就影响代码结构。
