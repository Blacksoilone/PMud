package world

// VerbHandler 处理一个动词的执行阶段。
// handler 应当读写 ctx，不直接操作 Action.Resp。
// 管线负责在 handler 返回后统一发送响应。
type VerbHandler func(l *Loop, ctx *AttemptContext)
