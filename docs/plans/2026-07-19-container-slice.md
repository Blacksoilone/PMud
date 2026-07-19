# 容器系统切片设计

## 设计原则

- 房间保持 `RoomItemLocation`，不做容器
- 玩家统一为容器，背包物品存在玩家"容器"里
- 箱子/收纳袋是独立容器物品
- 容器不可嵌套 carryable container

## 三种位置

```go
type ItemLocation interface{ isItemLocation() }

type RoomItemLocation struct {
    RoomID RoomID
}

type ContainerItemLocation struct {
    ContainerID string // "player:<PlayerID>" 或 "item:<ItemID>"
}
```

删除 `InventoryItemLocation`。玩家背包用 `ContainerItemLocation{ContainerID: "player:<id>"}`。

容器 ID 辅助函数：
```go
func PlayerContainerID(pid PlayerID) string { return "player:" + string(pid) }
func ItemContainerID(iid ItemID) string     { return "item:" + string(iid) }
```

## World 新增

### 容器状态（持久化，可变）

```go
type World struct {
    // ... 现有字段
    containerOpen   map[ItemID]bool    // open=true, closed=false（默认 false=关闭）
    containerLocks  map[ItemID]bool    // locked=true（默认 false=未锁）
}
```

- 容器物品默认是"关闭"状态
- `open` / `close` 切换状态
- 关闭时 look/examine 不显示容器内容
- 锁定状态由 `tag.lockable` 控制（已有）

### 操作

```go
// 玩家容器内容
func (w *World) itemsInContainer(containerID string) []ItemID

// 统一 move: item 从一个位置到另一个位置
func (w *World) moveItemToRoom(itemID ItemID, roomID RoomID) bool
func (w *World) moveItemToContainer(itemID ItemID, containerID string) bool

// 查询
func (w *World) containerIsOpen(itemID ItemID) bool
func (w *World) openContainer(itemID ItemID) bool
func (w *World) closeContainer(itemID ItemID) bool
func (w *World) containerContents(containerID string) []ItemID          // 前提：是打开的
func (w *World) containerRemainingCapacity(itemID ItemID) int

// Put / Get / Drop 统一
func (w *World) PutItem(itemID ItemID, targetContainerID string, playerID PlayerID) bool
func (w *World) GetItemFromContainer(containerID string, itemID ItemID, playerID PlayerID) bool
```

## 容器规则验证

```go
func (w *World) canHoldContainer(targetContainerID string) bool {
    // 如果 target is a player → always can （玩家可以背任何容器）
    // 如果 target is a carryable container（如收纳袋）→ cannot hold any container
    // 如果 target is a non-carryable container（如箱子）→ can hold carryable containers
}
```

容量验证：`tag.container` 的 `capacity` 参数控制最大物品数。

## session 命令层改动

需要新增/修改的命令：

| 命令 | 处理 |
|------|------|
| `open <item>` | 新命令 → 提交 OpenContainer 动作 |
| `close <item>` | 新命令 → 提交 CloseContainer 动作 |
| `put <item> in <container>` | 新命令 → 提交 PutItem 动作 |
| `get <item> from <container>` | 新命令 → 提交 GetFromContainer 动作 |
| `examine <item>` | 修改 → 如果是容器且打开，显示内容 |
| `inventory` | 无变化（内容仍来自 player container） |
| `drop <item>` | 无变化（仍从 player container → room） |
| `get <item>` | 无变化（仍从 room → player container） |

## 现有函数迁移

| 现函数 | 迁移后 |
|--------|--------|
| `itemsInInventory(playerID)` | `itemsInContainer(PlayerContainerID(playerID))` |
| `Inventory(playerID)` | `itemsInContainer(...)` 的 name wrapper |
| `InventoryItemIDs(playerID)` | `itemsInContainer(...)` |
| `GetItem(roomID, itemID, playerID)` | `moveItemToContainer(itemID, PlayerContainerID(playerID))` |
| `DropItem(roomID, itemID)` | `moveItemToRoom(itemID, roomID)` |
| `DropInventoryItem(roomID, itemID, playerID)` | 先检查在 player container 中，然后 moveItemToRoom |

## 测试计划

1. **位置迁移** — InventoryItemLocation → ContainerItemLocation，现有行为不变
2. **open/close** — 容器关闭时内容不可见，打开后可操作
3. **put/get** — 玩家和箱子间交互
4. **容量限制** — 达到 capacity 后 put 失败
5. **容器嵌套限制** — carryable container 拒绝其他 container 放入
6. **player 作为容器** — inventory 读取 player container

## 切片边界

**包含：**
- ✅ ContainerItemLocation 类型 + 删除 InventoryItemLocation
- ✅ 容器状态（open/close）
- ✅ itemsInContainer() 替代 itemsInInventory()
- ✅ 命令 open/close/put/get from
- ✅ 容量检查
- ✅ 容器嵌套限制
- ✅ 现有物品位置迁移（玩家背包）

**不包含：**
- ❌ 内容 source 层新增容器物品（后面再添加）
- ❌ lockable + container 联动（已有）
- ❌ TUI 特别渲染容器内容（初步用 look/list 显示）
