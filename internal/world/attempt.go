package world

import "PMud/internal/presentation"

// AttemptContext 贯穿整个 action 管线的上下文。
// Resolver、handler、hook 都读写这个结构。
type AttemptContext struct {
	PlayerID PlayerID
	Verb     string
	Input    string // 原始目标（物品短语、方向名等）
	World    *World // 运行时世界引用，供钩子使用

	// Resolve 阶段填充
	TargetItemID ItemID // 解析后的物品ID
	Direction    string // 规范化的方向

	// Execute 阶段填充
	Events  []presentation.Event
	NewRoom RoomID

	// 用于广播
	OldRoom  RoomID
	LeaveDir string

	// 控制字段
	Blocked     bool
	BlockReason string
}
