# PMud 开发路线图

## 已完成

- 基础世界模型（房间、物品、出口、锁定）
- 物品系统（拿/放/看，标签系统：carryable / lightable / container / lockable / exit）
- 双任务并行引擎（进度跟踪、弹窗选择）
- 「教程任务」+「练剑任务」
- 内容管线（source.json → compiler → world snapshot）
- 终端 TUI（房间视图、小地图、背包、任务面板）
- 简易锁定机制（铁门 + 旧油灯钥匙）

---

## 候选方向（按推荐顺序）

### 1. 容器
tag.container 已定义但未实现。玩家可以打开箱子，往里放/取东西。

- `open <item>` / `close <item>`
- `put <item> in <container>` / `get <item> from <container>`
- 容器有容量限制，可锁定
- 容器内物品暴露在 look 中（非隐藏容器），或需要搜索（隐藏容器）

### 2. 内容动词 — 可绑定逻辑
目前 content verb 只能发消息。需要将动词 handler 也放入内容管线。

- `pull_lever` / `push_button` / `insert_key` — 触发世界状态变化
- 动词 handler 在 go 代码中注册，内容数据按 id 引用
- 与任务系统联动：pull lever → trigger → quest advance

### 3. 任务全链路（内容定义 → 运行）
任务已经在 source.json 中有数据结构，但 compiler → server → progression 的管线还没闭合。

- 在 source.json 中定义 quest/stage/condition
- compiler 产出 ServerSnapshot.Quests
- progressionDefinitionsFromSnapshot 正确加载
- 端到端测试：写一个 content quest → 游戏中能推进

### 4. 持久化
退出后玩家状态不丢。

- bolt / sqlite 本地存储
- 保存：背包物品、任务进度、玩家位置
- 加载：重新进入世界时恢复
- 存档机制：自动存档 / 手动存档

### 5. NPC + 对话
可交互角色。

- NPC 在房间中可见
- `talk <npc>` / `ask <npc> about <topic>`
- 对话树在内容数据中定义
- NPC 可以给予物品、解锁出口、推进任务
- 简单的 AI：巡逻、交易、战斗（可选）

### 6. unlock / lock 动词
当前只能通过 tag 静态检查，缺少玩家的主动操作。

- `unlock <door> with <key>` / `lock <door>`
- 钥匙物品消耗或保留
- 与容器锁统一

### 7. 房间详细观察
- `look <direction>` 观察方向远景
- `look at <item>` 的更丰富描述（分第一次看/第二次看）
- 长描述 vs 短描述

### 8. 装备系统
- 武器/护甲 tag
- 装备栏（武器、防具、饰品槽位）
- `wield` / `wear` / `remove` / `unwield`
- 装备属性（攻击力、防御力）
- 装备条件（等级、属性要求）

### 9. 战斗系统
- `attack` / `kill` / `fight` 动词
- 回合制或即时制
- 原子属性（HP、攻击、防御、速度）
- 不同武器不同攻击方式（远程、近战、法术）
- 怪物 AI（仇恨、巡逻、逃跑）
- 战斗日志与动画反馈

### 10. 死亡系统
- 玩家 HP 归零 → 死亡
- 复活点（起始房间 / 绑定教堂 / 其他）
- 死亡惩罚（经验损失 / 物品掉落 / 金钱损失）
- 尸体容器（死亡后物品留在尸体中，可捡尸）
- 怪物死亡 → 尸体 + 可搜刮战利品
- 墓碑系统

### 11. 物品刷新与掉落
- 房间物品定时刷新（矿石、草药、普通装备）
- 怪物掉落表（概率 + 数量 + 条件）
- 稀有 / 唯一物品不刷新
- 刷新配置在内容数据中定义
- `respawn` / `repop` 调试命令

### 9. 地图与移动增强
- 区域切换（大地图 → 小地图）
- 跟随/组队
- climb / swim / fly 等特殊移动

### 10. 性能与工程
- 大型世界下的房间/物品索引
- 事件流去重/合并（批量推进不刷屏）
- 世界状态快照 → 增量更新
- 行为测试覆盖率提升

---

**决策原则**：

- 优先做**玩家能直接感知**的功能（容器、动词联动、NPC）
- 基础设施（持久化、性能）在需要时才做，不提前做
- 每个功能落地时优先保证内容管线可用（data + code 分离）
- 新增系统先通过 TUI 可验证，再考虑终端兼容
