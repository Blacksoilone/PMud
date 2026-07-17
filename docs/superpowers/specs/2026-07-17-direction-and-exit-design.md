其它mud是怎么做的？不仅仅是如何实现的，还有：它们的方位能被改变吗？它们的方位能被观察吗？

查了 TBAMUD、Evennia 和 CoffeeMud 后，发现大致有三种路线。

TBAMUD 这类 Diku/Circle 系把方向做成房间的固定槽位：north 槽里放目标房间、描述、门的关键词、关闭/上锁/隐藏状态。方向本身可以被 look，隐藏和锁定也可以改变，目标房间也能由脚本运行时修改；但它不是独立物品。门的名字只是这个方向槽位的交互关键词，移动仍然以 north 为主。

Evennia 更接近对象模型：出口就是一个对象，有名字、别名、destination、锁和 traversal hook。north、door、portal、jump out the window 在出口层没有本质区别，都是可以被命名和触发的出口对象。destination 可以运行时改变或清空；两个房间双向连接时就是两个独立出口对象。

CoffeeMud 是混合模型：房间仍按方向保存连接，但每个方向还可以有一个可观察、可开关、可上锁的 Exit 对象。它既能按方向查找，也能按名称查找。方向是空间入口，Exit 是玩家可见的通路表现。

对我们最有价值的不是“应该选对象还是槽位”，而是这几个事实：方向的目标可以改变；方向/出口可以被观察；隐藏和锁定不是静态地图属性；命令入口和玩家看到的名称可以分离。经典 MUD 往往把 north 作为导航词，把铁门作为这个方向上的交互名称；对象化 MUD 则把 north 或 portal 都视为可命名的出口对象。

这也说明“方位能不能被改变”至少有两种含义：改变 north 这个空间方向的名称，或者改变 north 当前指向的房间。前者通常不允许，因为 north 是稳定的命令和空间语义；后者很常见，可以用于移动房间、传送、机关、建造和脚本。

“方位能不能被观察”也不应简单回答是或否。经典系统观察的是方向上的描述、门和状态；对象系统观察的是出口对象本身。玩家实际需要的是：输入 look north 能得到这个方向当前可见的描述、是否有门/阻挡、是否能通过，以及必要时它通向哪里，而不一定需要知道内部结构。

这对本项目的提醒是：不要过早把“方向”“出口”“门”“传送门”强行合并或强行拆开。应该分别确认三个层次：玩家用什么找到它，世界如何保存它的目标，哪些 Tag 改变它的可见性和通行性。现有 Tag 设计适合表达动态状态，但还不能仅凭其它 MUD 的实现直接决定我们的承载形式。

参考：

- TBAMUD 方向槽位与出口状态：https://github.com/tbamud/tbamud/blob/03d7ba7a48495b270e113821bc59b3cbde43c4b0/src/structs.h#L778-L800
- TBAMUD 运行时修改出口：https://github.com/tbamud/tbamud/blob/03d7ba7a48495b270e113821bc59b3cbde43c4b0/src/dg_objcmd.c#L605-L696
- Evennia 出口对象：https://github.com/evennia/evennia/blob/0c677ae652422db397519ee80afc6cf2d6f52c2b/docs/source/Components/Exits.md#L20-L38
- Evennia 动态 destination：https://github.com/evennia/evennia/blob/0c677ae652422db397519ee80afc6cf2d6f52c2b/evennia/commands/default/building.py#L1240-L1339
- CoffeeMud Exit 接口：https://github.com/bozimmerman/CoffeeMud/blob/5e4f8aac3aa1c8a5490c0906ef1cff4d27fd6e5d/com/planet_ink/coffee_mud/Exits/interfaces/Exit.java#L34-L109

名称约定确实还可以用在别处，但最好只用于“名字和语义天然一一对应”的地方，并且由编译器识别，运行时不要反复猜字符串。

最明显的是标准方位。物品的规范内部名是 north，且拥有 exit tag，编译器就把它识别为 north 方位；铁门的名字不是 north，所以永远不会成为 north。名称和身份在这里恰好是同一件事，额外 direction 字段反而可能制造矛盾。

内容文本键也适合使用约定。比如物品 ID 是 item.tutorial.lantern，可以默认推导 item.tutorial.lantern.name、.inner_name、.description，作者只在需要例外时覆盖。房间、任务、阶段也可以类似减少重复字段。这个约定只是在编译期寻找默认文本，不应改变对象行为。

Tag 名本身也已经是一种名称协议：exit、hidden、lock 由代码中的 TagDefinition 注册。数据写出这个名字，编译器找到对应定义并检查 schema，之后运行时使用编译好的定义或 ID，而不是 switch 字符串。

命令语言中的保留词也适合，例如 self、here、all、north。这些名称不是普通别名，而是解析器明确认识的语法词。它们应被集中登记，避免内容作者意外创建同名对象后产生不确定语义。

但任务阶段顺序、所属关系、奖励、条件之类不适合只靠名称推导。即使 stage ID 带有 quest 前缀，也不应因此自动决定它属于哪个任务或下一阶段是谁，因为重命名会悄悄改变游戏逻辑。名字可以提供默认值，关系和状态机仍要显式声明。

大概可以形成一条原则：名称约定适合消除“重复描述同一个事实”的字段，不适合替代真正的关系和规则。编译器应拒绝约定和显式数据冲突，并把结果转成类型化快照；这样名称的简洁只存在于作者层，不会把魔法字符串扩散到世界运行逻辑里。
