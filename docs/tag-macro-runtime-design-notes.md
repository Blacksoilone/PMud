# Tag Macro and Runtime Tag Design Notes

## 结论

`tag` 是装备和实体机制的核心底层，但作者面对的 tag 与运行时规则读取的 tag 不必是同一层。

采用三层模型：

- `source_tags`：创作者写在内容里的宏标签，可以是复合标签、语义快捷方式或材料概念。
- `tag definitions`：代码里的 tag 行为定义，每个 definition 只写一次，声明参数 schema 和响应哪些规则上下文。
- `tag instances`：编译后挂在实体或 part 上的数据实例，只包含 tag id 和参数；运行时规则、检查、效果系统实际读取它们。

例如创作者可以写：

```json
{
  "source_tags": ["silver"]
}
```

编译后得到：

```json
{
  "tag_instances": {
    "visual.silver": {},
    "effect.undead_slayer": { "multiplier": 1.5 },
    "property.soft": { "hardness": 3 }
  }
}
```

运行时不需要理解 `silver` 这个创作期概念。它只读取 `visual.silver`、`effect.undead_slayer`、`property.soft` 这些 tag instance，并调用同名 tag definition 的代码。

物品数据里不会出现任何可执行代码。数据只声明“我调用这个 tag definition，并传入这些参数”。即使某个效果只被一个物体使用一次，也必须写成 tag definition 代码，再由物体上的 tag instance 调用它。

## 为什么这样做

如果让 `silver` 同时承担描述、亡灵增伤、护甲惩罚、耐久损耗等含义，它会变成一个语义大包，后续很难维护。

如果让作者手动写：

```json
["silver", "undead_slayer", "soft"]
```

则依赖作者自觉维护语义一致性，内容规模变大后很容易漏写或写错。

复合 tag 编译能把复杂度收束到编译期：

- 多层展开是确定性图问题。
- 循环依赖可以编译时报错。
- 重复标签可以去重。
- 输出可以稳定，方便调试和 diff。
- 运行时只面对底层 tag instance 集合。

这比在运行时解释模糊语义更可控。

## 命名隔离

tag 不决定名字，也不生成 alias。

一个物品可以机制上拥有：

```json
{
  "tag_instances": {
    "visual.silver": {},
    "effect.undead_slayer": { "multiplier": 1.5 },
    "property.soft": { "hardness": 3 },
    "effect.evil_slayer": { "multiplier": 1.25 }
  }
}
```

但它的名字仍然可以是：

```json
{
  "name_key": "item.toilet_brush.name",
  "aliases": ["matongshua", "toilet_brush", "马桶刷"]
}
```

名字、alias、主描述由作者显式决定。程序不因为底层 tag 推导“它应该叫银剑”或“它应该有 jinjian 这个输入名”。

## Tag Definition 和 Tag Instance

`TagDefinition` 是代码，不是内容数据。它包含：

- 唯一 tag id，例如 `effect.undead_slayer`；
- 参数 schema、默认值和合法范围；
- 允许挂载的作用域，例如 item 根、part 或 actor；
- 响应的规则上下文，例如 attack、examine、equip；
- 实际执行行为的 Go 代码，未来也可以是 Lua 代码，但仍然属于代码层。

`TagInstance` 是内容数据。它只包含：

- tag id；
- 传给对应 `TagDefinition` 的参数。

例如：

```json
{
  "tag_instances": {
    "effect.undead_slayer": {
      "multiplier": 1.5
    }
  }
}
```

这等价于一次受约束的函数调用：

```text
effect.undead_slayer.Apply(ctx, { multiplier: 1.5 })
```

但数据里不写函数体、不写表达式、不写脚本。行为代码只在 tag definition 文件中出现一次。

## 描述标签

运行时可以保留专门给观察系统使用的描述性标签，例如：

```text
visual.silver
visual.holy
visual.burned
```

这些标签可以参与 `examine` 时的补充观察，但不应该改变实体身份或输入名。

也就是说，`silver` 作为 source tag 不保留到运行时；保留的是 `visual.silver` 这个 tag instance，并由 `visual.silver` 的 tag definition 决定它在观察上下文中如何表现。

## Part 的定位

`part` 是 tag 的作用域，也是玩家可以理解、观察和推理的玩法语义。

第一版不计划提供自由拆装、替换或组合 part 的玩家操作，但这不代表 part 对玩家不可见。玩家当然可以理解：银质剑刃对亡灵有效，银质刀把则不该产生同样效果。

例如：

```json
{
  "parts": {
    "blade": {
      "source_tags": ["silver"]
    },
    "hilt": {
      "source_tags": ["sacred_wood"]
    }
  }
}
```

编译后：

```json
{
  "parts": {
    "blade": {
      "tag_instances": {
        "visual.silver": {},
        "effect.undead_slayer": { "multiplier": 1.5 },
        "property.soft": { "hardness": 3 }
      }
    },
    "hilt": {
      "tag_instances": {
        "visual.sacred_wood": {},
        "effect.evil_slayer": { "multiplier": 1.25 }
      }
    }
  }
}
```

part tag 不自动冒泡到 item 根。规则系统必须明确声明自己读取哪个作用域：根、指定 part、任意 part，或某类 part。

例如，对亡灵增伤的规则应该读取攻击接触部位：

```text
weapon.part(blade).has(effect.undead_slayer)
```

而不是读取整个 item 是否“某处有银”。如果只有 hilt 拥有 `visual.silver` 或 `effect.undead_slayer`，它不应该让剑刃攻击亡灵时获得银质增伤。

## 合成、贴附不是第一版玩法

合成、贴附不作为第一版玩家玩法系统。

part 仍然可以作为玩家理解物品机制的语义结构存在：它帮助规则表达“哪个部位带来哪个效果”，也帮助观察系统解释“你看到的效果来自哪里”。只是第一版不要求玩家自由拆装、替换、组合物品。

未来如果加入合成或贴附，优先把它们视为内容层扩展：

- 通过数据定义哪些组合存在；
- 通过 tag 检查判断是否满足条件；
- 通过已有 tag definition 和 tag instance 规则产生效果；
- 不因为加入合成而改动整体实体框架。

换句话说，未来合成或贴附应尽量是“内容使用 tag 系统”，而不是“重新设计物品框架”。

## 展开规则

source tag catalog 支持多层展开：

```json
{
  "silver": [
    { "id": "visual.silver", "data": {} },
    { "id": "effect.undead_slayer", "data": { "multiplier": 1.5 } },
    { "id": "property.soft", "data": { "hardness": 3 } }
  ],
  "holy": [
    { "id": "visual.holy", "data": {} },
    { "id": "effect.evil_slayer", "data": { "multiplier": 1.25 } }
  ],
  "holy_silver": [
    "silver",
    "holy"
  ]
}
```

`holy_silver` 最终展开为：

```text
visual.silver {}
effect.undead_slayer { multiplier: 1.5 }
property.soft { hardness: 3 }
visual.holy {}
effect.evil_slayer { multiplier: 1.25 }
```

编译器需要处理：

- 多层递归展开；
- 循环依赖检测；
- 未定义 tag 报错；
- 重复 tag instance 去重或合并；
- 稳定输出顺序。

tag instance 默认仍然是集合语义：同一作用域上同一个 tag id 不靠重复出现表达叠加。输出顺序不参与规则判断，只用于可读性、调试、序列化稳定性和错误信息稳定性。

当前建议采用：

- 按作者写入的 `source_tags` 顺序展开；
- 每个 source tag 按 catalog 中 `expands_to` 顺序展开；
- 重复 tag instance 去重，保留第一次出现位置；如果重复实例携带不同参数，编译器应报错，除非该 tag definition 明确声明合并规则。

这个顺序策略不是核心语义，未来可以调整为字典序或 namespace 分组，不应影响玩法规则。

## 重复和叠加

tag instances 默认是集合语义。重复出现不叠加。

如果未来需要可叠加数值效果，不应该靠重复 tag 表达，而应该由对应 tag definition 明确声明可重复或合并规则，例如：

```json
{
  "id": "modifier.damage_bonus",
  "data": {
    "target": "undead",
    "value": 0.1,
    "stacking": "additive"
  }
}
```

这样可以避免 tag 系统偷偷变成隐式数值系统：是否能叠加、如何叠加，必须由 tag definition 的代码和 schema 明确声明。

## 调度规则

运行时不应该写中心化分支：

```go
if tag == "effect.undead_slayer" {
    // undead slayer behavior
}
```

它应该通过 registry 和 dispatcher 调用 tag definition：

```text
for each relevant TagInstance in action context:
  definition = registry.Lookup(instance.ID)
  definition.Handle(context, instance.Data)
```

dispatcher 只负责找到当前上下文相关的 instance、查询 registry、调用对应 hook。它不理解“银”“亡灵杀手”“柔软材料”的具体规则。具体规则属于各自 tag definition 文件。

## 第一版范围

第一版不需要完整 tag macro compiler，也不需要 tag rule engine。

第一版核心仍然是：

- 数据驱动世界；
- 名字、描述、alias；
- 房间、物品、背包、移动；
- `look` / `examine` / `get` / `drop` / `inventory`；
- 一个可试玩的 tutorial loop。

tag macro 和 runtime tag rule engine 是后续系统，但当前文档确定方向：

```text
创作期 source tag -> 编译期展开和 schema 校验 -> 运行时 tag instance -> 调用代码中的 tag definition
```

不要让运行时直接解释模糊的 source tag，不要让 tag 自动决定名字和 alias，也不要让内容数据包含任何可执行逻辑。
