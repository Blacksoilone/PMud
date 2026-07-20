# dsv4系统性修复计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** 修复 dsv4 最近提交中会影响后续系统正确性的 Tag、出口、Progression、协议和容器边界问题，并把稳定设计写给后续 DeepSeek 实现者。

**Architecture:** 保留 dsv4 的 ActionAttempt 和 Tag hook 方向，不重构整个 Loop。内容编译期严格解析注册 Tag；世界动作按 action domain 解析对象；Progression 用明确 quest ID 和显式 lifecycle 结算；协议使用可逆编码；容器取物复用现有体积约束。

**Tech Stack:** Go、现有 content/world/session/progression/presentation/protocol/TUI 包、标准测试、race/shuffle、golangci-lint。

## Tasks

- [x] `internal/content`: 未注册 Tag、未知字段、缺失/错误参数在编译期失败；保留合法 Tag 的类型化实例。
- [x] `internal/world/loop.go`: 让无方向具名出口通过实际 `move` resolver 进入 ActionAttempt，并保持方向出口行为不变。
- [x] `internal/progression`: 将 reward resolution 绑定 quest ID，补齐 lifecycle/activation 的最小权威语义，保持 stage machine 不变。
- [x] `internal/presentation` / `internal/protocol` / `internal/client`: 将 quest_list 改为可逆的结构化 wire 编码，覆盖分隔符文本。
- [x] `internal/world/container.go`: 从容器取物前检查玩家体积硬上限；重量仍允许超重。
- [x] 全量验证：targeted tests、LSP diagnostics、`go test -race -shuffle=on -count=1 ./...`、lint、diff check。
