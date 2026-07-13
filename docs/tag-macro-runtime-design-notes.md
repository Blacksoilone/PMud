# Tag Macro and Runtime Tag Design Notes

## 结论

`tag` 是装备和实体机制的核心底层，但作者面对的 tag 与运行时规则读取的 tag 不必是同一层。

采用两层模型：

- `source_tags`：创作者写在内容里的标签，可以是复合标签、语义快捷方式或材料概念。
- `runtime_tags`：编译后得到的底层标签集合，运行时规则、检查、效果系统实际读取它们。

例如创作者可以写：

```json
{
  "source_tags": ["silver"]
}
```

编译后得到：

```json
{
  "runtime_tags": [
    "visual.silver",
    "effect.undead_slayer",
    "property.soft"
  ]
}
```

运行时不需要理解 `silver` 这个创作期概念。它只读取 `visual.silver`、`effect.undead_slayer`、`property.soft` 这些底层标签。

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
- 运行时只面对底层标签集合。

这比在运行时解释模糊语义更可控。

## 命名隔离

tag 不决定名字，也不生成 alias。

一个物品可以机制上拥有：

```json
{
  "runtime_tags": [
    "visual.silver",
    "effect.undead_slayer",
    "property.soft",
    "effect.evil_slayer"
  ]
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

## 描述标签

运行时可以保留专门给观察系统使用的描述性标签，例如：

```text
visual.silver
visual.holy
visual.burned
```

这些标签可以参与 `examine` 时的补充观察，但不应该改变实体身份或输入名。

也就是说，`silver` 作为 source tag 不保留到运行时；保留的是 `visual.silver`。

## Part 的定位

`part` 是 tag 的作用域，不是玩家必须理解的核心玩法。

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
      "runtime_tags": [
        "visual.silver",
        "effect.undead_slayer",
        "property.soft"
      ]
    },
    "hilt": {
      "runtime_tags": [
        "visual.sacred_wood",
        "effect.evil_slayer"
      ]
    }
  }
}
```

part tag 不自动冒泡到 item 根。规则系统必须明确声明自己读取哪个作用域：根、指定 part、任意 part，或某类 part。

## 合成、贴附和 Part 不是第一版玩法

合成、贴附、part 不作为第一版玩家玩法系统。

它们可以作为创作者组织内容、表达作用域和支撑 tag 规则的底层能力存在，但第一版不要求玩家自由拆装、替换、组合物品。

未来如果加入合成或贴附，优先把它们视为内容层扩展：

- 通过数据定义哪些组合存在；
- 通过 tag 检查判断是否满足条件；
- 通过已有 runtime tag 规则产生效果；
- 不因为加入合成而改动整体实体框架。

换句话说，未来合成或贴附应尽量是“内容使用 tag 系统”，而不是“重新设计物品框架”。

## 展开规则

source tag catalog 支持多层展开：

```json
{
  "silver": [
    "visual.silver",
    "effect.undead_slayer",
    "property.soft"
  ],
  "holy": [
    "visual.holy",
    "effect.evil_slayer"
  ],
  "holy_silver": [
    "silver",
    "holy"
  ]
}
```

`holy_silver` 最终展开为：

```text
visual.silver
effect.undead_slayer
property.soft
visual.holy
effect.evil_slayer
```

编译器需要处理：

- 多层递归展开；
- 循环依赖检测；
- 未定义 tag 报错；
- 重复 runtime tag 去重；
- 稳定输出顺序。

runtime tag 的语义是集合。输出顺序不参与规则判断，只用于可读性、调试、序列化稳定性和错误信息稳定性。

当前建议采用：

- 按作者写入的 `source_tags` 顺序展开；
- 每个 source tag 按 catalog 中 `expands_to` 顺序展开；
- 重复 tag 去重，保留第一次出现位置。

这个顺序策略不是核心语义，未来可以调整为字典序或 namespace 分组，不应影响玩法规则。

## 重复和叠加

runtime tags 默认是集合语义。重复出现不叠加。

如果未来需要可叠加数值效果，不应该靠重复 tag 表达，而应该使用显式 effect 或 modifier 数据结构，例如：

```json
{
  "effect": "damage_bonus",
  "target": "undead",
  "value": 0.1
}
```

这样可以避免 tag 系统偷偷变成数值系统。

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
创作期复合 tag -> 编译期展开 -> 运行时底层 tag 集合
```

不要让运行时直接解释模糊的 source tag，也不要让 tag 自动决定名字和 alias。
