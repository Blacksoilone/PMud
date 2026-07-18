# Tag 注册系统切片设计

## 目标

将硬编码的 `Tag` 结构体替换为 `TagDefinition` + `TagInstance` 注册系统。实现后：
- 新增 tag 类型只需注册定义 + 写行为代码，不改 Item 结构体
- Tag 实例带参数（例如 exit 的目标房间、container 的容量）
- 内容数据编译时自动校验 tag 合法性

## 核心类型

### `world/tag.go` — Tag 定义与实例

```go
type TagID string // "tag.exit", "tag.carryable", "tag.container"

type TagScope string
const (
    TagScopeItem TagScope = "item"
    TagScopeRoom TagScope = "room"
)

type TagFieldType string
const (
    TagFieldString TagFieldType = "string"
    TagFieldBool   TagFieldType = "bool"
    TagFieldInt    TagFieldType = "int"
    TagFieldRef    TagFieldType = "ref" // 引用 item/room ID
)

type TagField struct {
    Name     string
    Type     TagFieldType
    Required bool
    Default  any
}

type TagDefinition struct {
    ID          TagID
    Description string
    Scopes      []TagScope
    Fields      []TagField
}

// TagInstance 替换当前的 Tag struct
type TagInstance struct {
    DefinitionID TagID
    Params       map[string]any // field name → value
}
```

### Item 结构体变更

```go
type Item struct {
    NameKey        string
    InnerName      string
    DescriptionKey string
    Name           string
    Description    string
    Aliases        []string
    Tags           []TagInstance  // 原来是 []Tag
}
```

删除旧的 `Tag`、`Exit` 结构体。

## World 变更

```go
type World struct {
    // ... 现有字段
    tagDefinitions map[TagID]TagDefinition  // 新增
}
```

### 注册函数

```go
func (w *World) RegisterTag(def TagDefinition) error {
    // 校验：ID 唯一、字段名唯一、scope 合法
    // 存入 w.tagDefinitions
}
```

### 内置 tag 定义（在 New() 中注册）

```go
tag.exit:      方向(可选), 目标房间(必填) → 当前 Exit 对应逻辑
tag.carryable: 无参数                → 当前 Carryable 对应逻辑
tag.container: 容量(可选, 默认1), 可打开(默认true) → 新功能
tag.lockable:  钥匙ID(可选)          → 新功能
```

**第一片只注册 tag.exit 和 tag.carryable，完全替换现在行为。** tag.container 和 tag.lockable 注册但不实现行为，留到后续切片。

## 行为代码迁移

当前直接检查 `tag.Exit` / `tag.Carryable` 字段的地方改为查询 TagInstance：

```go
// 旧
func (w *World) itemExit(itemID ItemID) (Exit, bool) {
    for _, tag := range item.Tags {
        if tag.Exit != nil {
            return *tag.Exit, true
        }
    }
    return Exit{}, false
}

// 新
func (w *World) itemExit(itemID ItemID) (Exit, bool) {
    item, ok := w.items[itemID]
    if !ok { return Exit{}, false }
    params, ok := item.tagParams("tag.exit")
    if !ok { return Exit{}, false }
    exit := Exit{
        Direction:    params["direction"].(string),  // 可能是 ""
        TargetRoomID: RoomID(params["target"].(string)),
    }
    return exit, true
}

// Item 辅助方法
func (i Item) tagParams(tagID TagID) (map[string]any, bool) {
    for _, inst := range i.Tags {
        if inst.DefinitionID == tagID {
            return inst.Params, true
        }
    }
    return nil, false
}
```

类似地 `itemIsCarryable`：

```go
func (w *World) itemIsCarryable(itemID ItemID) bool {
    item, ok := w.items[itemID]
    if !ok { return false }
    _, ok = item.tagParams("tag.carryable")
    return ok
}
```

## 内容编译层变更

### `content/types.go`

```go
// ServerTag 不变：
type ServerTag struct {
    Carryable bool
    Exit      *ServerExit
}

// 新增 SourceTag 类型用于未来 CUE 数据
// 当前 ServerTag → worldTags() 改为生成 TagInstance
```

### `world/constructors.go`

```go
func worldTags(tags []content.ServerTag) []TagInstance {
    result := make([]TagInstance, 0, len(tags))
    for _, tag := range tags {
        if tag.Carryable {
            result = append(result, TagInstance{
                DefinitionID: "tag.carryable",
                Params:       map[string]any{},
            })
        }
        if tag.Exit != nil {
            params := map[string]any{"target": string(tag.Exit.TargetRoomID)}
            if tag.Exit.Direction != "" {
                params["direction"] = tag.Exit.Direction
            }
            result = append(result, TagInstance{
                DefinitionID: "tag.exit",
                Params:       params,
            })
        }
    }
    return result
}
```

## 删除

- `world/types.go`: 删除 `Tag` 结构体、`Exit` 结构体
- 所有 `tag.Exit` / `tag.Carryable` 字段引用改为 `tagParams("tag.exit")` / `tagParams("tag.carryable")`

## 测试

- 现有行为测试全部不变（行为相同，只是内部实现换了）
- 新增：
  - `TestTagRegistry_defineAndLookup` — 注册定义后能查到
  - `TestTagRegistry_rejectsDuplicateID` — 重复注册抛错
  - `TestTagInstance_Params` — 参数读写正确
  - `TestTagBackwardCompat` — 现有物品 tag 通过新系统仍能正确解析 exit/carryable

## 切片边界

**这个 slice 做：**
- ✅ `TagDefinition` / `TagInstance` / `TagField` 类型
- ✅ `World.RegisterTag()` + `tagDefinitions` map
- ✅ 删除旧 `Tag` / `Exit` 结构体
- ✅ 修改 `Item.Tags` 为 `[]TagInstance`
- ✅ 迁移 `itemExit` / `itemIsCarryable` 等全部 tag 访问点
- ✅ 内置 `tag.exit` / `tag.carryable` 注册
- ✅ `worldTags()` 编译转换
- ✅ 全量测试通过

**这个 slice 不做：**
- ❌ 不实现 tag.container / tag.lockable 的行为
- ❌ 不引入行为 hook 系统（TagDefinition 还没有 callback/handler）
- ❌ 不改动 content source 格式（ServerTag 保留）
- ❌ 不改动 TUI / wire protocol

## 文件变更清单

| 文件 | 变更 |
|------|------|
| `world/tag.go` | 新增：TagDefinition / TagInstance / RegisterTag |
| `world/types.go` | 删除 Tag / Exit；Item.Tags 改为 []TagInstance |
| `world/items.go` | itemExit / itemIsCarryable / ordinaryItemsInRoom / DropItem 迁移 |
| `world/navigation.go` | exitItemIDs（依赖 itemExit）连带迁移 |
| `world/constructors.go` | New() 注册内置 tag；worldTags() 输出 TagInstance |
| `world/loop.go` | 无变化（只依赖 itemExit / itemIsCarryable 等访问器） |
| `world/world_test.go` | 新增注册测试；现有测试不变 |
| `content/types.go` | 无变化（ServerTag 保留） |
