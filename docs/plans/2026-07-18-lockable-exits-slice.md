# Lockable 退出门切片设计

## 目标

添加 `tag.lockable`，让退出门可以被锁住，玩家需要持有指定钥匙才能通过。

## 设计

### 新 tag 定义

```
tag.lockable
  字段: key_item_id (ref, 必填) — 能打开此锁的物品 ID
```

### 编译层

```go
type LockableTag struct {
    KeyItemID ItemID  // "item.tutorial.old_lantern"
}

// ServerTag 新增:
type ServerTag struct {
    Exit      *ExitTag
    Carryable bool
    Lightable bool
    Container *ContainerTag
    Lockable  *LockableTag  // 新增
}
```

`SourceTag` 格式：
```json
{"ID": "lockable", "Params": {"key_item_id": "item.tutorial.old_lantern"}}
```

### World 行为

`MovePlayer` 增加锁检查：

```
for each exit in currentRoom:
    if exit.Direction matches direction:
        if exit has tag.lockable:
            if player does NOT have key_item_id in inventory:
                return (currentRoom, false, "locked")
        return (exit.TargetRoomID, true, "")
return (currentRoom, false, "")
```

`MovePlayer` 返回类型改为 `(RoomID, bool, string)`，string 是失败原因（`""` 表示成功，`"locked"` 表示锁住）。

### Loop 处理

```go
nextRoom, ok, reason := world.MovePlayer(event.PlayerID, event.Direction)
if !ok {
    if reason == "locked" {
        w.SendToPlayer(event.PlayerID, systemMessage("system.move.locked"))
    } else {
        w.SendToPlayer(event.PlayerID, systemMessage("system.move.blocked"))
    }
    continue
}
```

### 教程内容更新

- 北方退出门（`item.tutorial.north`）添加 `tag.lockable`，钥匙是旧油灯
- 旧油灯保持现有 `tag.carryable` + `tag.lightable`
- 添加新文本 `system.move.locked`

玩家体验：
1. 进入入口，看到北方通道
2. 往北走 → "门锁着。"（因为没拿油灯）
3. 拿起旧油灯（完成 get_lantern 任务阶段）
4. 往北走 → 用油灯开门，进入练习场（完成 enter_yard 任务阶段）
5. 旧油灯既是门钥匙又是 quest 目标，一物两用

### 文件变更

| 文件 | 变更 |
|------|------|
| `world/tag.go` | 注册 tag.lockable 定义 |
| `world/types.go` | 无（TagInstance 已支持） |
| `world/navigation.go` | MovePlayer 改为 (RoomID, bool, string) + 锁检查 |
| `world/player.go` | MovePlayer 改为新签名 |
| `world/loop.go` | move 处理添加 locked 分支 |
| `world/events.go` | 无 |
| `world/world_test.go` | 新增锁门测试，更新 MovePlayer 调用 |
| `content/types.go` | +LockableTag + 常量 |
| `content/compiler.go` | +lockable 编译 |
| `content/fixture.go` | north 加锁 |
| `data/tutorial/source.json` | north 加锁 |
| `world/constructors.go` | north 加锁 TagInstance |

### 测试

- `TestLockableExitBlocksWithoutKey` — 无钥匙不能通过
- `TestLockableExitAllowsWithKey` — 有钥匙可以通过
- `TestLockableExit_compilePipeline` — SourceTag → ServerTag → TagInstance

### 不做

- ❌ `tag.lockable` 的 `pickable` 字段（可用其他钥匙开锁）
- ❌ 破坏/砸锁
- ❌ 锁的可见性（look 时显示"锁着的"）
- ❌ 多钥匙/组合锁
