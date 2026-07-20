# DeepSeek设计澄清

这份文档用于给后续实现者说明当前项目中已经确定、不能被简化模型覆盖的设计边界。

## Tag系统

- `TagDefinition` 是代码：负责稳定 ID、参数 schema、默认值、作用域和 hook。
- `TagInstance` 是数据：只保存 tag ID 和参数。
- `SourceTag` 是作者层宏，可以在编译期展开成一个或多个 TagInstance。
- 未注册的 tag 必须在内容编译期报错，不能作为 GenericTag 静默透传。
- 内容数据不能包含可执行逻辑；条件、奖励和行为由代码注册并执行。
- Part 是明确的 tag scope。Part 上的 tag 不自动冒泡到物品根节点。

当前 `exit`、`carryable`、`container`、`lightable` 等实现可以作为迁移对象，但不能继续增加中央 `switch tag.ID` 分支。

## 出口与物品

- 出口是带 `exit` Tag 的房间内物品。
- `exit` 的目标房间是显式参数；标准方向可以由规范内部名在编译期识别，不需要额外 direction 字段。
- `north` 就是 north，铁门就是铁门；不能把铁门建模成 north，也不能把方位作为铁门属性。
- 无方向的传送门、洞口等仍可通过名称解析为出口。
- `exit` 与 `carryable` 正交。通常 north 不带 carryable，但不能在编译器中强制禁止 `exit + carryable`。
- 体积是拾取硬上限；重量可以超重，超重状态阻止移动。容器取物也必须遵守体积硬上限。
- 同一房间同一标准方向最多一个出口；编译期和运行时放置都要维护该不变量。
- 服务端解析对象短语；客户端不能解析对象名称、别名或 ID。

## Progression

- Quest 是单一 current_stage 的 stage machine，不是 DAG。
- Stage 没有独立 lifecycle；finish_conditions 全部满足才推进。
- 最终 stage 先进入 `reward_pending`，奖励处理完成后才进入 `completed` 或 `waiting_refresh`。
- lifecycle 需要支持 `hidden`、`unlocked`、`active`、`reward_pending`、`completed`、`waiting_refresh`、`retry_wait`。
- 激活策略只有 `manual_accept`、`auto_on_event`、`always_active`；不存在 `auto_on_condition`。
- 奖励结算必须绑定明确 quest ID，不能从所有 reward_pending 任务中按排序挑第一个。
- repeatable quest 使用 refresh_at；玩家存档保存运行时状态和摘要，不保存定义。
- 时间/调度层在 `refresh_at` 到期后调用 `Engine.RefreshQuest(player_id, quest_id)`；Progression 引擎不自行读取系统时间。刷新会清空阶段进度并按激活策略回到 `active`、`unlocked` 或 `hidden`。
- 复杂可选/并行内容使用独立 quest，不加入原生 optional objective 或并行轨道。

## 协议和TUI

- 服务端事件字段必须可安全承载任意内容文本，不能依赖未转义的逗号、竖线等分隔符。
- TUI 是产品界面；弹窗覆盖内容区但保留输入栏。
- 小地图当前只显示当前房间的一跳八方向邻居，不推导邻居之间的连线。

## 本轮修复边界

本轮修复 dsv4 已确认的问题：未知 Tag 编译失败、具名特殊出口进入实际动作管线、quest reward 按 ID 结算、progression lifecycle 与激活策略、quest_list 协议编码、容器取物体积校验。暂不重构大型 Loop 文件，也不实现战斗、NPC、多人系统。
