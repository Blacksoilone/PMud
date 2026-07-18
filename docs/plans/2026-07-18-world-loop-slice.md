# World Loop 切片设计

## 核心变化

引入一个专用 goroutine（World Loop）作为 `*World` 的唯一写入者。所有需要访问 world state 的操作都通过 channel 提交，串行执行。读操作（look、examine 等）先走 loop，后续可以优化为旁路。

```
TCP A ─→ session A ─→  world.Queue(Action)  ─→┐
TCP B ─→ session B ─→  world.Queue(Action)  ─→┤
TCP C ─→ session C ─→  world.Queue(Action)  ─→┤
                                                ↓
                                    World Loop (single goroutine)
                                    owns: *World, *Hub, *progression.Engine
                                                │
                           ┌────────────────────┼────────────────────┐
                           ↓                    ↓                    ↓
                     return events        broadcast to room    keepalive/tick
                     to session A         via Hub             (future)
```

## 新增

### `world/loop.go`

```go
// Action 是 session 向 loop 提交的最小工作单元。
type Action struct {
    PlayerID PlayerID
    Verb     string   // "move" | "get" | "drop" | "examine" | "look" | "inventory" | "quest"
    Target   string   // direction, item ID, or empty
    Resp     chan<- ActionResult
}

type ActionResult struct {
    Events      []presentation.Event  // 返回给提交者的 events
    NewRoom     RoomID                // session 更新 currentRoom 用
}

// Queue 是 session 调用的入口（非阻塞）。
func (l *Loop) Queue(a Action) { l.ch <- a }
```

### `world/hub.go`

```go
type SessionHub struct {
    sessions map[PlayerID]chan<- []presentation.Event
}

func (h *SessionHub) Register(id PlayerID, ch chan<- []presentation.Event)
func (h *SessionHub) Unregister(id PlayerID)
func (h *SessionHub) SendToRoom(roomID RoomID, events []presentation.Event, exclude PlayerID)
```

## 修改

### `session/session.go` — 主要重构

1. `sessionState` 不再持有 `*World`，改为持有 `*world.Loop`
2. `handleLine` 不再直接调 world 方法，改为通过 `Loop.Queue` 提交 Action
3. 每个 session 有一个 `incoming chan []presentation.Event` loop 可以推广播
4. session goroutine：select 读取 loop 返回 + incoming channel 的广播

**伪代码：**

```go
type sessionState struct {
    loop        *world.Loop
    incoming    chan []presentation.Event  // loop 推广播到这里
    currentRoom world.RoomID
    playerID    world.PlayerID
}

func (s *sessionState) handleLine(line string) []presentation.Event {
    parsed := parse(line)
    switch cmd := parsed.(type) {
    case command.MoveCommand:
        resp := make(chan world.ActionResult, 1)
        s.loop.Queue(world.Action{
            PlayerID: s.playerID,
            Verb:     "move",
            Target:   cmd.Direction,
            Resp:     resp,
        })
        result := <-resp
        s.currentRoom = result.NewRoom
        return result.Events

    case command.ItemCommand:
        // 类似：pack Verb/Target/PlayerID → Queue → wait Resp
    }
}
```

### `world/loop.go` — 主循环

```go
func (l *Loop) run() {
    for a := range l.ch {
        switch a.Verb {
        case "move":
            nextRoom, ok := l.world.MovePlayer(a.PlayerID, a.Target)
            if !ok {
                a.Resp <- blockResult()
                continue
            }
            playerRoom := l.world.SetPlayerRoom(a.PlayerID, nextRoom)
            events := l.roomObservationEvent(nextRoom)

            // 广播给同房间其他玩家
            l.hub.SendToRoom(nextRoom, roomEnterEvent(a.PlayerID), a.PlayerID)
            // 通知旧房间玩家离开
            l.hub.SendToRoom(...)

            // 检查 progression
            events = append(events, l.applyProgression(a.PlayerID, progression.Trigger{Kind: TriggerMovedRoom})...)

            a.Resp <- ActionResult{Events: events, NewRoom: nextRoom}

        case "get":
            // ...
        }
    }
}
```

### `server.go`

- 在 `StartSession` 中创建并启动 `world.Loop`（`go loop.Run()`）

## 切片边界

**这个 slice 只做：**
- ❌ 不重构 progression（只把 engine 从 session 移到 loop）
- ❌ 不做 keepalive、NPC、计时器（留钩子）
- ❌ 不改变 TUI 客户端
- ❌ 不改变 wire protocol

**这个 slice 做：**
- ✅ 新增 `world.Loop`（Action channel + run goroutine）
- ✅ 新增 `world.SessionHub`（注册/广播）
- ✅ `session.handleLine` 改为 queue + wait
- ✅ 所有 world 写操作走 loop
- ✅ 玩家进出房间的广播通知
- ✅ loop 内 progression 处理

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `world/loop.go` | 新增 |
| `world/hub.go` | 新增 |
| `world/types.go` | World 结构体无变化（Loop/Hub 在外部引用） |
| `session/session.go` | handleLine 改为 Queue+wait，新增 incoming channel |
| `session/server.go` | StartSession 创建并启动 Loop |
| `session/session_test.go` | 更新 newTestSessionState |
| `world/loop_test.go` | 新增 loop 测试 |
