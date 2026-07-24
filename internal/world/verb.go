package world

// VerbHandler 处理一个动词的执行阶段。
// handler 应当读写 ctx，不直接操作 Action.Resp。
// 管线负责在 handler 返回后统一发送响应。
type VerbHandler func(l *Loop, ctx *AttemptContext)

// ItemResolver 返回指定 action 上下文相关的物品列表。
// 管线根据这些物品的 tag 实例来执行 hook（前置检查/后置事件）。
// 如果某个动词没有注册 resolver，使用默认 fallback（扫描全屋+背包中 hook.Verbs 匹配的物品）。
type ItemResolver func(l *Loop, ctx *AttemptContext) []*Entity

// VerbSource 标记动词的来源。
type VerbSource string

const (
	VerbSourceBuiltin     VerbSource = "builtin"      // Go 代码内置
	VerbSourceContent     VerbSource = "content"      // 内容数据声明
	VerbSourceHookRefOnly VerbSource = "hook_ref"     // 仅被 tag hook 引用，无独立声明
)

// VerbEntry 是 Loop 动词注册表中的一条记录。
type VerbEntry struct {
	Name       string
	Source     VerbSource
	MessageKey string // content 声明的默认成功消息
}
