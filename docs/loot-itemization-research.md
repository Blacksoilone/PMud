# 装备与随机词缀系统调研笔记

这份笔记总结对 Diablo、Wynncraft 以及若干 ARPG/MMO 装备系统的调研，目标不是复刻它们，而是提炼“如何让随机装备有惊喜但不失控”的设计原则。

结论先行：成熟的 loot game 很少做“从全体效果里随便抽几个”的完全随机。它们几乎都在做 **受控随机**：先确定 base/template/source，再在受限词缀池里随机数值或少量效果；真正改变玩法的效果通常由设计者手工定义，或被严格限制在特定模板、稀有度、来源、插槽、职业、等级、制作系统和互斥组里。

## 总体结论

我们之前讨论的方向是对的：

```text
模板白名单
+ 积分预算
+ 独占组
+ 风险标签
+ 生成后审计
+ 来源/掉落等级
+ 持久世界风险控制
```

这个组合比单纯积分制更接近成熟游戏的做法。

不同游戏使用的术语不同，但本质相似：

- Diablo II 用 base item、item quality、prefix/suffix、affix level、affix group、treasure class、unique/set/runeword/crafted 分层控制随机。
- Diablo III 用 smart loot、legendary/set power、Mystic 单项重铸、Kanai's Cube 等方式把 build-defining power 从普通随机词缀里分离出来。
- Diablo IV 近年的 itemization 改版把普通词缀数量压低，把复杂度移动到 legendary aspect、tempering、masterworking、greater affix 等受限系统里。
- Wynncraft 的普通装备更接近“手工模板 + bounded identification rolls”，Major ID 是手工设计的 build-defining effect，不是任意随机词缀。
- Path of Exile、Last Epoch、Grim Dawn 等游戏也大量使用 base tag、slot、prefix/suffix、tier、exclusive family、crafting potential、drop-only tier、monster-specific base 等边界。

对我们的 MUD 来说，最重要的一句话是：

> 随机生成应该随机“模板内允许的变化”，而不是随机“世界中所有可能的规则”。

## Diablo II：多阶段受控随机

Diablo II 的装备系统很适合学习“为什么随机不应该是自由组合”。它不是先随机一堆效果，而是多阶段生成：

```text
掉落来源 / Treasure Class
  -> base item
  -> item level / quality level
  -> item quality
  -> quality-specific generator
  -> affix / socket / unique / set / runeword / crafted
```

### Magic 与 Rare

Magic item 最多：

```text
1 prefix + 1 suffix
```

Rare item 最多：

```text
3 prefixes + 3 suffixes
```

并且不能重复同一 affix group。这是非常重要的：随机词缀不是无限堆叠，而是有位置、数量和互斥约束。

可借鉴为：

```text
普通随机物品:
  1 minor prefix + 1 minor suffix

稀有随机物品:
  最多 2-3 个普通属性
  最多 1 个 major 属性
  同一 exclusive group 不重复
```

### Unique、Set、Runeword

Diablo II 的强身份物品大多不是随机词缀堆出来的。

```text
Unique:
  手工定义身份和属性组合

Set:
  手工定义套装与件数奖励

Runeword:
  玩家把 exact rune sequence 放入 exact socket base，得到手工定义的强效果
```

这说明：build-defining effect 往往需要设计者手工控制。系统可以允许玩家通过 runeword 这类机制“制造强物品”，但配方本身仍是设计者定义的。

### Normal base 的价值

D2 中普通白装并不只是垃圾，因为 runeword 需要非魔法、有正确孔数和类型的 base。这是一个很重要的经济设计：

```text
基础物品本身有价值
不是所有价值都来自随机词缀
```

对 MUD 的启发：

- 一根特殊材质的棍子，即使没有魔法词缀，也可以因合成/贴附/符文/仪式系统而有价值。
- 随机词缀不应该吞掉 base/template/source 的存在感。
- base item 可以决定允许哪些 tag、slot、major effect 和制作路线。

### Ethereal 的启发

D2 的 ethereal 提供强数值，但通常不可修理。这是典型的强度-约束交易：

```text
更强基础数值
换取耐久风险和生命周期限制
```

对 MUD 来说，这对应：

```text
强效果可以存在，但要绑定 durability、charges、腐化、不可修理、不可交易、失控风险等约束。
```

## Diablo III：传奇效果成为 build 核心

Diablo III 后期装备系统的核心不是随机 rare，而是 legendary/set effects。

普通属性提供数值，真正改变玩法的是：

```text
legendary power
set bonus
Kanai's Cube extracted power
```

Mystic 重铸也很受限：通常只能重铸一项属性，并锁定其他属性。这说明制作系统不应该让玩家随意重写整件物品，否则随机掉落和物品身份都会失去意义。

可借鉴为：

```text
随机装备:
  允许微调一项属性
  不允许完全重写身份

强效果:
  可萃取/转移时必须按槽位或类别受限
```

比如如果 MUD 将来允许“抽取 tag/effect”，也应该有槽位限制：

```text
武器效果槽
防具效果槽
饰品效果槽
环境/房间效果槽
```

不能让所有强效果随便贴到任何东西上。

## Diablo IV：把复杂度移出普通词缀

Diablo IV 的规则随赛季和版本变化很频繁，所以不能把某个版本细节当成永久事实。但近年改版方向很清楚：减少普通掉落词缀噪音，把强化路径拆成多个受限系统。

典型层次：

```text
普通 affixes
legendary aspect / Codex
tempering
masterworking
greater affix
unique / mythic unique
```

### Legendary Aspect

legendary aspect 是 build-defining effect 的承载方式之一。它不是普通数值词缀，而是按类别、装备槽位和 Codex 系统管理。

对 MUD 的启发：

```text
major effect 应该进入专门层，而不是和 +3 damage 放在同一个随机池。
```

### Tempering 与 Masterworking

Tempering 使用 manual/category 限制可加的属性类型；Masterworking 提供后期强化，并在特定层级强化某项 affix。

这说明制作不是“无限自由编辑”，而是：

```text
在受限类别中增加一层变化
通过有限次数/材料/失败风险控制上限
```

对 MUD 来说，未来可以有：

```text
磨砺: 只影响武器物理 stats
灌注: 只影响元素/气味/温度/诅咒类 tags
铭刻: 只影响 trigger/effect 槽
淬炼: 有次数或耐久成本
```

### Greater Affix

Greater Affix 的设计价值在于“高信号掉落”：玩家一眼知道这件东西有特殊潜力，而不是从一堆低价值随机数里筛垃圾。

对文字 MUD 来说，也应该避免让玩家读大量无意义随机属性。强随机应该有可见信号：

```text
这把剑的锋刃异常完美。
这件斗篷散发着不自然的寒意。
这枚戒指的铭文似乎仍在移动。
```

## Wynncraft：手工模板 + bounded rolls

Wynncraft 对我们很有价值，因为它是多人持久世界，而且装备高度依赖 build。

它的装备大多不是完全随机生成，而是：

```text
手工物品模板
  + identification 数值范围
  + requirements
  + elements
  + powder slots
  + major IDs
```

### Identifications

Identifications 是装备上的数值属性，可以有正负范围。玩家鉴定后看到 roll 结果。

这非常适合 MUD：

```text
模板定义某属性可能出现的范围
实例保存具体 roll
```

不是：

```text
从所有属性里随便抽
```

而是：

```text
这件物品本来就设计为有这些属性，只是数值随机。
```

### Major ID

Major ID 是 build-defining 规则变化，通常特定物品拥有，不是普通随机池乱滚。

这和我们讨论的“major 属性 / 独占组”高度一致：

```text
minor stats:
  可以 roll 数值

major effects:
  手工放在特定模板或特殊掉落里
```

### Powder Slots

粉末系统提供玩家控制的后处理：给装备加入元素倾向或特殊效果。它不是完全重写物品，而是在槽位和元素规则内定制。

对 MUD 的启发：

```text
贴附/灌注/镶嵌系统可以作为玩家控制随机之外的路线。
```

比如：

```text
一把铁剑随机生成后，玩家可以用火焰粉末、寒霜油脂、毒液结晶进一步定向修改。
```

但这些修改也必须受槽位、冲突组和风险审计控制。

### 正负属性共存

Wynncraft 很多装备通过正负属性塑造身份。强物品不只是数值高，而是有取舍。

这对 MUD 很重要：

```text
强效果不一定靠高成本禁止
也可以靠负面 tags、负面 stats、使用条件、耐久风险、环境限制来平衡
```

## 其他 ARPG 的共同模式

### Path of Exile

PoE 的 modifier 系统高度复杂，但几个原则很清楚：

```text
base item tags 决定可出现 modifier
modifier 有 prefix/suffix、domain、generation type、level、spawn weight
modifier family/group 防止互斥属性重复
influence、corruption、enchant 等是额外层
```

可借鉴：

```text
affix 不只是名字和值，还需要：
  slot
  family
  tier
  min_level
  allowed_tags
  blocked_tags
  spawn_weight
  value_range
```

### Last Epoch

Last Epoch 的装备通常最多：

```text
2 prefixes + 2 suffixes
```

T6/T7 是 drop-only 高阶词缀，crafting 受 Forging Potential 限制。Unique 可以通过 Eternity Cache 与 Exalted affixes 结合成 Legendary。

启发：

- 高阶强属性可以只允许掉落，不允许普通制作生成。
- 制作需要资源/潜力上限，防止无限修正。
- Unique identity 可以和少量随机高阶属性融合，但融合过程必须受限。

### Grim Dawn

Grim Dawn 使用 base + prefix/suffix，并有 Monster Infrequent：特定怪物掉落的主题 base，再叠加随机词缀。

启发：

```text
来源身份很重要。
```

一把“火龙掉落的剑”不应该只是在全局池里抽出的剑。它可以有：

```text
source: 火龙
allowed tags: 火焰、龙族、高温、燃烧、鳞甲、威压
blocked tags: 水生、寒霜、普通木质
```

这比全局随机更有叙事感，也更好平衡。

### Borderlands

Borderlands 的枪械由厂商、武器类型、部件、元素和来源共同塑造。它的启发是：随机生成可以很丰富，但必须有“品牌/来源/类型”这些模板化边界。

对 MUD 来说，类似：

```text
锻造流派
地区
怪物来源
材料来源
工匠签名
宗派
诅咒体系
```

这些都可以成为随机边界。

## 对我们的 MUD 的设计映射

### 1. 分离 Routine Stats 和 Build-Defining Effects

普通数值和改变玩法的效果不应该在同一个池子里。

```text
Routine Stats:
  damage
  sharpness
  weight
  durability
  resistance
  fuel_amount
  burn_intensity
  accuracy

Build-Defining Effects:
  on_hit 变青蛙
  on_attack_hit 放逐
  on_damage_taken 冻结攻击者
  suppress_drop
  duplicate_reward
  reset_instance
```

Routine stats 可以较频繁随机。Build-defining effects 需要模板白名单、独占组、风险标签和审计。

### 2. 随机应从 SourceContext 开始

推荐生成流程：

```text
SourceContext
  掉落来源、地区、怪物、箱子、副本、任务、事件

BaseTemplate
  物品类型、slot、材质、等级、archetype

Routine Stat Packages
  数值包、roll ranges、minor tags

Optional Build-Effect Template
  major effect / unique-like effect

Audit
  检查预算、独占组、风险闭环、持久污染
```

这比“先随机一个铁剑，再从全局 tag 池抽”更好。

### 3. 物品模板应该声明可随机空间

每个模板应声明：

```text
base tags
required components
allowed minor affixes
allowed major effect groups
forbidden effects
exclusive groups
stat ranges
risk budget
source restrictions
trade policy
persistence policy
```

例如：

```text
飞镖模板:
  base tags: [飞镖, 投掷物, 可消耗]
  allowed major effects:
    - 中毒
    - 变形
    - 穿甲
    - 燃烧
  default constraints:
    - limited_charges
    - not_repairable
```

这样“能把对手变成青蛙的飞镖”可以存在，但它天然有消耗性边界。

### 4. 日常掉落牺牲上限是合理的

Diablo、Wynncraft 和其他 ARPG 都说明：日常掉落不需要覆盖所有可能的疯狂组合。真正破格的东西应来自特殊来源。

推荐：

```text
普通野外掉落:
  routine stats + 少量 minor tags

普通副本非特殊产出:
  routine stats + 最多 1 个 major group

Boss 特殊产出:
  可有手工 major effect

手工神器:
  可突破普通约束，但必须显式设计

异常/错误技式生成:
  可越界，但应有不稳定、不可交易、副本内限定、短生命周期等约束
```

### 5. 强物品需要生命周期和经济约束

持久 MUD 比单局 ARPG 更容易被装备污染。

所以强物品除了战斗强度，还要评估：

```text
是否可交易
是否可修理
是否可复制
是否可长期保存
是否可继承
是否能影响他人
是否能影响经济
是否能改写世界状态
```

比如：

```text
一次性青蛙飞镖:
  可能可以

永久可交易青蛙戒指:
  高风险

副本内临时青蛙法杖:
  可能可以
```

### 6. 来源身份可以替代全局随机

不要让所有地方掉落同一套随机池。掉落来源应该塑造可出现的属性。

```text
火龙:
  火焰、龙族、高温、威压、燃烧、鳞片

下水道:
  腥味、滑溜溜、毒、潮湿、腐败、狭窄

古代图书馆:
  记忆、文字、诅咒、静默、知识、纸质、灰尘
```

这让随机装备更像世界的一部分，而不是抽卡机。

## 推荐写入系统设计的规则

### 装备生成管线

```text
1. 选择 SourceContext
2. 选择 BaseTemplate
3. 按模板和来源过滤 affix/effect 池
4. Roll routine stats
5. 按来源等级决定是否 roll major effect
6. 检查 exclusive groups
7. 检查 risk tags 与 required containers
8. 检查 persistence/economy policy
9. 生成实例并保存 provenance
```

### Affix / Effect 元数据

每个随机属性不应该只是名字和值，而应有：

```text
id
kind: prefix/suffix/implicit/major/effect
family / exclusive_group
tier
min_level
allowed_slots
allowed_base_tags
blocked_base_tags
source_tags
spawn_weight
value_range
risk_tags
requires_container
trade_policy
persistence_risk
```

### 何时手工设计 Unique

当一个效果涉及这些内容时，应倾向手工设计，而不是进入普通随机池：

```text
改变技能语义
改变资源循环
改变移动/空间/出口规则
新增 trigger
影响多人互动
影响经济
影响持久世界状态
跨多个系统产生连锁反应
```

## 对第一版的影响

第一版仍然不应该实现完整随机装备系统。

但这次调研强化了一个底层要求：未来的物品系统需要保存 provenance 和 generation metadata。

即使第一版只有手写物品，也最好在概念上允许：

```text
definition_id
instance_id
source_context
template_id
rolled_stats
attached_effects
trade_policy
persistence_policy
```

第一版可以先只实现其中很小一部分，但不要把物品写成没有来源、没有实例、没有未来扩展空间的静态结构。

## 参考来源类型

本次调研主要参考了以下类型资料：

- Blizzard / Arreat Summit 的 Diablo II item、magic item、crafted item、runeword、socketed item 资料
- Diablo II Resurrected 社区数据表与 reverse engineering 资料，例如 TreasureClassEx、MagicPrefix、MagicSuffix、UniqueItems、SetItems、Runes
- Blizzard Diablo III patch notes、Mystic、Kanai's Cube 相关官方说明与社区整理
- Blizzard Diablo IV Loot Reborn、Codex、Tempering、Masterworking、Greater Affix 等官方和社区资料
- Wynncraft 官方 API 文档、item database、set metadata、identifications、powders、elements、crafting、identifier、mythic/legendary/set 资料
- Path of Exile 官方 item data / mod data、PoE Wiki modifier/crafting 资料
- Last Epoch 官方 support 资料：affixes、crafting、rarity、implicits、Legendary/Eternity Cache
- Grim Dawn 官方 loot guide 与 patch notes
- Borderlands / Lootlemon、Guild Wars 2 itemstats 等辅助参考

这些资料的共同方向非常一致：**好装备系统不是随机性越大越好，而是随机发生在精心设计的边界之内。**
