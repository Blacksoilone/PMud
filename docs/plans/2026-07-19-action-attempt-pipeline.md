# ActionAttempt 管线设计

## 痛点

当前 `loop.go` 中，每个动词的处理是独立硬编码的：

```
handle(a Action) → switch a.Verb → handleMove / handleGet / ...
```

每个 handler 自行处理：
- 参数解析
- 物品名消歧
- tag 条件判断（lockable 检查内嵌在 `MovePlayer` 中）
- 广播事件
- progression trigger 触发

**问题**：
1. 加行为 = 加 `case` + 加 `handleXxx` + 可能加 World 方法
2. Tag 无法声明"我挂钩什么事件"
3. Tag 行为逻辑散落在 World 方法和 Loop 处理函数中
4. `progression.Trigger` 是平行机制，每次手动调用 `applyProgression`

## 目标

一个统一的管线，让动词和 tag 行为通过注册加入，不用改 `loop.go`。

## 设计

### 核心概念

```
Action → Pipeline → ActionResult
```

管线阶段：

```
1. Resolve   — 解析目标（物品名 → 物品ID，方向别名 → 规范方向）
2. PreHook   — tag 前置钩子（"可以这么做吗？"）可阻断
3. Execute   — 动词 handler 执行业务逻辑
4. PostHook  — tag 后置钩子（添加响应事件，触发 progression）
5. Respond   — 组装最终事件 + 广播
```

### 类型定义

```go
// Action 保持不变
type Action struct {
    PlayerID PlayerID
    Verb     string
    Target   string
    Resp     chan<- ActionResult
}

// 管线上下文 —— 贯穿整个 pipeline
type AttemptContext struct {
    PlayerID PlayerID
    Verb     string
    Input    string          // 原始输入目标

    // Resolve 阶段填充
    TargetItemID  ItemID     // 解析后的物品ID（如果有）
    TargetRoomID  RoomID
    Direction     string

    // Execute 阶段填充
    Events        []presentation.Event  // 主响应事件
    NewRoom       RoomID                // 玩家新房间
    ProgressionTriggers []progression.Trigger

    // 控制字段
    Blocked       bool
    BlockReason   string    // "locked", "blocked" 等
}
```

### 动词注册

```go
type VerbHandler func(ctx *AttemptContext) error

func (l *Loop) RegisterVerb(verb string, handler VerbHandler)

// 使用
loop.RegisterVerb("move", func(ctx *AttemptContext) error {
    // 检查 exit，处理 lockable
    nextRoom, ok, reason := l.world.MovePlayer(ctx.PlayerID, ctx.Direction)
    if !ok {
        ctx.Blocked = true
        ctx.BlockReason = reason
        return nil  // 不是错误，是被阻断
    }
    ctx.NewRoom = nextRoom
    obs, _ := l.world.Look(nextRoom)
    ctx.Events = append(ctx.Events, newRoomObservationEvent(obs))
    return nil
})
```

### Tag 钩子注册

Tag 可以声明挂钩哪些阶段、哪些动词：

```go
type HookPhase int
const (
    HookPreAction  HookPhase = iota  // 执行前检查
    HookPostAction                    // 执行后附加
)

type TagHook struct {
    Phase   HookPhase
    Verbs   []string     // 空 = 所有动词
    Handler func(ctx *AttemptContext, params map[string]any) error
}

// TagDefinition 新增字段
type TagDefinition struct {
    ID          TagID
    Description string
    Scopes      []TagScope
    Fields      []TagField
    Hooks       []TagHook       // ← 新增
}
```

### 管线流程

```
func (l *Loop) handleAction(a Action) {
    ctx := &AttemptContext{
        PlayerID: a.PlayerID,
        Verb:     a.Verb,
        Input:    a.Target,
    }

    // 1. Resolve
    l.resolveAttempt(ctx)

    // 2. Pre-hooks
    for _, hook := range l.hooksFor(ctx.Verb, HookPreAction) {
        hook(ctx)
        if ctx.Blocked { break }
    }

    // 3. Execute
    if !ctx.Blocked {
        if handler, ok := l.verbHandlers[ctx.Verb]; ok {
            handler(ctx)
        }
    }

    // 4. Post-hooks
    for _, hook := range l.hooksFor(ctx.Verb, HookPostAction) {
        hook(ctx)
    }

    // 5. Progression triggers
    for _, t := range ctx.ProgressionTriggers {
        ctx.Events = append(ctx.Events, l.applyProgression(ctx.PlayerID, t)...)
    }

    // 6. Respond
    if ctx.Blocked {
        a.Resp <- blockedResult(ctx)
    } else {
        a.Resp <- ActionResult{Events: ctx.Events, NewRoom: ctx.NewRoom}
    }

    // 7. Broadcast
    l.broadcastRoomChange(ctx)
}
```

### 现有动词迁移

示例 — `move` 动词迁移为使用管线：

```go
// 注册 handler
loop.RegisterVerb("move", moveHandler)

func moveHandler(ctx *AttemptContext) error {
    oldRoom := l.world.PlayerCurrentRoom(ctx.PlayerID)
    nextRoom, ok, reason := l.world.MovePlayer(ctx.PlayerID, ctx.Direction)
    if !ok {
        ctx.Blocked = true
        ctx.BlockReason = reason
        return nil
    }
    ctx.NewRoom = nextRoom
    obs, _ := l.world.Look(nextRoom)
    ctx.Events = append(ctx.Events, newRoomObservationEvent(obs))
    ctx.ProgressionTriggers = append(ctx.ProgressionTriggers,
        progression.Trigger{Kind: progression.TriggerMovedRoom, RoomID: string(nextRoom)})
    if oldRoom != nextRoom {
        ctx.LeftRoom = oldRoom
        ctx.Direction = ctx.Direction
    }
    return nil
}
```

`tag.lockable` 注册前置钩子：

```go
// 在 RegisterTag 内部或独立加载
tag.lockable.GetDefinition().Hooks = []TagHook{
    {Phase: HookPreAction, Verbs: []string{"move"}, Handler: lockableHook},
}

func lockableHook(ctx *AttemptContext, params map[string]any) error {
    keyID, _ := params["key_item_id"].(string)
    if keyID == "" { return nil }
    if !l.world.PlayerHasItem(ctx.PlayerID, ItemID(keyID)) {
        ctx.Blocked = true
        ctx.BlockReason = "locked"
    }
    return nil
}
```

### 动词生命周期

| 阶段 | 可阻断 | 可修改上下文 |
|------|--------|-------------|
| Resolve | 否 | 是 |
| Pre-hook | 是 | 是（可修改 Target 等） |
| Execute | 否（handler 自己设 Blocked） | 是 |
| Post-hook | 否 | 是（可追加事件） |

### 迁移策略

不一次性重写全部 handler。逐步迁移：

1. 先建好管线框架 + `RegisterVerb` + `RegisterHook`
2. 迁移 `move`（含 lockable hook）
3. 迁移 `get`（含 container pre-check）
4. 迁移其余动词
5. 删除旧的 `handleMove` / `handleGet` / `switch a.Verb` 分支
6. 最后迁移 `_enter_world` 和 `_leave_world`（系统动词）

### 非目标

- ❌ 不改变 `Action` / `ActionResult` 类型（接口稳定）
- ❌ 不改变 World API 签名
- ❌ 不改变 `progression.Engine` API
- ❌ 不改 session 层（sessions 只管发命令、收事件）

## 文件清单

| 文件 | 变更 |
|------|------|
| `world/attempt.go` | 新增：AttemptContext, Resolver, Pipeline Run |
| `world/verb.go` | 新增：VerbHandler, RegisterVerb, verbHandlers map |
| `world/hook.go` | 新增：TagHook, HookPhase, hook 注册 |
| `world/tag.go` | 修改：TagDefinition 加 Hook 字段 |
| `world/loop.go` | 修改：handleAction 使用管线，逐个迁移 handler |
| `world/player.go` | 修改：PlayerHasItem 改为导出方法 |
| `world/look.go` | 新增：广播逻辑从 handler 中抽出 |

## 后续

管线完成后：

- `tag.container` → 注册 `get` 的 pre-hook（从容器取物需要先打开容器）
- `tag.lightable` → 注册 `light` 动词 handler + `look` 的 post-hook（点亮后房间描述不同）
- `tag.combatant` → 注册 `attack` 动词 handler
