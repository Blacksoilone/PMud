package world

import (
	"log"
	"sync"

	"PMud/internal/presentation"
	"PMud/internal/progression"
)

func (l *Loop) registerBuiltinVerbs() {
	l.RegisterVerb("_enter_world", handleEnterWorld)
	l.RegisterVerb("_leave_world", handleLeaveWorld)
	l.RegisterVerb("move", handleMove)
	l.RegisterVerb("look", handleLook)
	l.RegisterVerb("get", handleGet)
	l.RegisterVerb("drop", handleDrop)
	l.RegisterVerb("examine", handleExamine)
	l.RegisterVerb("look-item", handleLookItem)
	l.RegisterVerb("inventory", handleInventory)
	l.RegisterVerb("quest", handleQuest)
	l.RegisterVerb("verb", handleVerbList)

	// 注册内置动词到注册表
	for name := range l.verbHandlers {
		if name == "" || name[0] == '_' {
			continue
		}
		l.verbRegistry[name] = VerbEntry{Name: name, Source: VerbSourceBuiltin}
	}

	// 物品解析器 — 管线用它们找到当前 action 相关的物品
	l.RegisterItemResolver("move", resolveMoveItem)
	l.RegisterItemResolver("get", resolveRoomItemByPhrase)
	l.RegisterItemResolver("examine", resolveVisibleItemByPhrase)
	l.RegisterItemResolver("look-item", resolveVisibleItemByPhrase)
	l.RegisterItemResolver("drop", resolveInventoryItemByPhrase)
}

// Verbs 返回当前注册表快照中的全部动词。
func (l *Loop) Verbs() map[string]VerbEntry {
	result := make(map[string]VerbEntry, len(l.verbRegistry))
	for k, v := range l.verbRegistry {
		result[k] = v
	}
	return result
}

// initContentVerbs 将内容声明的动词合并到注册表中。
// 对尚无 handler 的 content 动词，自动注册默认 handler。
func (l *Loop) initContentVerbs(verbs map[string]VerbEntry) {
	for name, entry := range verbs {
		if _, exists := l.verbRegistry[name]; exists {
			continue
		}
		l.verbRegistry[name] = entry
		if _, hasHandler := l.verbHandlers[name]; !hasHandler {
			l.RegisterVerb(name, contentVerbDefaultHandler)
		}
	}
}

// validateVerbRefs 扫描所有 tag definitions 的 hook.Verbs 引用，
// 标记那些不在注册表中的动词为 hook_ref，并打印警告。
func (l *Loop) validateVerbRefs() {
	for _, def := range l.world.tagDefinitions {
		for _, hook := range def.Hooks {
			for _, verb := range hook.Verbs {
				if verb == "" || verb[0] == '_' {
					continue
				}
				if _, exists := l.verbRegistry[verb]; !exists {
					l.verbRegistry[verb] = VerbEntry{Name: verb, Source: VerbSourceHookRefOnly}
					log.Printf("warning: verb %q referenced by tag hook (%q) has no handler or content declaration", verb, def.ID)
				}
			}
		}
	}
}

// contentVerbDefaultHandler 是内容声明动词的默认 handler。
// 它发送该动词的默认成功消息，让 tag hooks 负责实际效果。
func contentVerbDefaultHandler(l *Loop, ctx *AttemptContext) {
	entry, ok := l.verbRegistry[ctx.Verb]
	if !ok || entry.MessageKey == "" {
		return
	}
	fields := map[string]string{"input": ctx.Input}
	if ctx.TargetItemID != "" {
		fields["item"] = string(ctx.TargetItemID)
	}
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: entry.MessageKey,
		Fields:     fields,
	})
}

// handleVerbList 列出当前世界全部可用动词。
func handleVerbList(l *Loop, ctx *AttemptContext) {
	var body string
	for name, entry := range l.verbRegistry {
		if name == "" || name[0] == '_' {
			continue
		}
		body += "  " + name
		switch entry.Source {
		case VerbSourceBuiltin:
			body += "  [内置]"
		case VerbSourceContent:
			body += "  [内容]"
		case VerbSourceHookRefOnly:
			body += "  [钩子引用]"
		}
		if entry.MessageKey != "" {
			body += "  " + entry.MessageKey
		}
		body += "\n"
	}
	if body == "" {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			Message: "没有可用动词。",
		})
		return
	}
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		Message: "可用动词:\n" + body,
	})
}

// resolveMoveItem 找出当前房间中匹配输入方向的那个退出门物品。
func resolveMoveItem(l *Loop, ctx *AttemptContext) []Item {
	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	for _, itemID := range l.world.exitItemIDs(roomID) {
		exit, ok := l.world.itemExit(itemID)
		if ok && exit.Direction == ctx.Input {
			return []Item{l.world.items[itemID]}
		}
	}
	return nil
}

// resolveRoomItemByPhrase 在当前房间中按短语匹配物品。
func resolveRoomItemByPhrase(l *Loop, ctx *AttemptContext) []Item {
	resolution := l.world.ResolveRoomItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.Input)
	if resolution.Found {
		if item, ok := l.world.items[resolution.ItemID]; ok {
			return []Item{item}
		}
	}
	return nil
}

// resolveVisibleItemByPhrase 在当前房间+玩家背包中按短语匹配物品。
func resolveVisibleItemByPhrase(l *Loop, ctx *AttemptContext) []Item {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, ctx.Input)
	if resolution.Found {
		if item, ok := l.world.items[resolution.ItemID]; ok {
			return []Item{item}
		}
	}
	return nil
}

// resolveInventoryItemByPhrase 在玩家背包中按短语匹配物品。
func resolveInventoryItemByPhrase(l *Loop, ctx *AttemptContext) []Item {
	resolution := l.world.ResolveInventoryItemPhrase(ctx.PlayerID, ctx.Input)
	if resolution.Found {
		if item, ok := l.world.items[resolution.ItemID]; ok {
			return []Item{item}
		}
	}
	return nil
}

type ActionResult struct {
	Events  []presentation.Event
	NewRoom RoomID
}

type Action struct {
	PlayerID PlayerID
	Verb     string
	Target   string
	Resp     chan<- ActionResult
}

type Loop struct {
	world       *World
	progression *progression.Engine
	actions     chan Action
	mu          sync.RWMutex
	sessions    map[PlayerID]chan<- []presentation.Event

	verbHandlers  map[string]VerbHandler
	itemResolvers map[string]ItemResolver
	verbRegistry  map[string]VerbEntry
}

func NewLoop(w *World) *Loop {
	l := &Loop{
		world:         w,
		progression:   progression.NewEngine(w.ProgressionDefinitions()),
		actions:       make(chan Action, 64),
		sessions:      make(map[PlayerID]chan<- []presentation.Event),
		verbHandlers:  make(map[string]VerbHandler),
		itemResolvers: make(map[string]ItemResolver),
		verbRegistry:  make(map[string]VerbEntry),
	}
	l.registerBuiltinVerbs()
	l.initContentVerbs(w.contentVerbs)
	l.validateVerbRefs()
	return l
}

func (l *Loop) Start() {
	go l.run()
}

func (l *Loop) Submit(a Action) {
	l.actions <- a
}

func (l *Loop) EnterWorld(playerID PlayerID) (RoomID, bool) {
	resp := make(chan ActionResult, 1)
	l.Submit(Action{PlayerID: playerID, Verb: "_enter_world", Resp: resp})
	result := <-resp
	return result.NewRoom, len(result.Events) == 0
}

func (l *Loop) LeaveWorld(playerID PlayerID) {
	resp := make(chan ActionResult, 1)
	l.Submit(Action{PlayerID: playerID, Verb: "_leave_world", Resp: resp})
	<-resp
}

func (l *Loop) Register(playerID PlayerID, outgoing chan<- []presentation.Event) {
	l.mu.Lock()
	l.sessions[playerID] = outgoing
	l.mu.Unlock()
}

func (l *Loop) Unregister(playerID PlayerID) {
	l.mu.Lock()
	delete(l.sessions, playerID)
	l.mu.Unlock()
}

func (l *Loop) RegisterVerb(verb string, handler VerbHandler) {
	l.verbHandlers[verb] = handler
}

func (l *Loop) RegisterItemResolver(verb string, resolver ItemResolver) {
	l.itemResolvers[verb] = resolver
}

func hookMatchesVerb(hook TagHook, verb string) bool {
	if len(hook.Verbs) == 0 {
		return true
	}
	for _, v := range hook.Verbs {
		if v == verb {
			return true
		}
	}
	return false
}

func (l *Loop) SendToRoom(roomID RoomID, events []presentation.Event, exclude PlayerID) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for _, pid := range l.world.PlayersInRoom(roomID) {
		if pid == exclude {
			continue
		}
		if ch, ok := l.sessions[pid]; ok {
			select {
			case ch <- events:
			default:
			}
		}
	}
}

func (l *Loop) run() {
	for a := range l.actions {
		l.handleWithRecovery(a)
	}
}

func (l *Loop) handleWithRecovery(a Action) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic handling action %+v: %v", a, r)
			if a.Resp != nil {
				a.Resp <- ActionResult{
					Events: []presentation.Event{
						presentation.SystemMessageEvent{MessageKey: "system.internal_error"},
					},
				}
			}
		}
	}()
	l.handle(a)
}

func (l *Loop) handle(a Action) {
	l.handleAction(a)
}

func (l *Loop) handleAction(a Action) {
	handler, ok := l.verbHandlers[a.Verb]
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.unknown_command"},
			},
		}
		return
	}

	ctx := &AttemptContext{
		PlayerID: a.PlayerID,
		Verb:     a.Verb,
		Input:    a.Target,
		World:    l.world,
		OldRoom:  l.world.PlayerCurrentRoom(a.PlayerID),
	}

	// 1. Pre-hooks — 从相关物品 tag 定义中找前置钩子
	items := l.relevantItems(ctx)
	for _, item := range items {
		for _, inst := range item.Tags {
			def, ok := l.world.TagDefinition(inst.DefinitionID)
			if !ok {
				continue
			}
			for _, hook := range def.Hooks {
				if hook.Phase != HookPreAction || !hookMatchesVerb(hook, a.Verb) {
					continue
				}
				hook.Handler(ctx, inst.Params)
				if ctx.Blocked {
					break
				}
			}
			if ctx.Blocked {
				break
			}
		}
		if ctx.Blocked {
			break
		}
	}

	// 2. Execute
	if !ctx.Blocked {
		handler(l, ctx)
	}

	// 3. Post-hooks — 从相关物品 tag 定义中找后置钩子
	for _, item := range l.relevantItems(ctx) {
		for _, inst := range item.Tags {
			def, ok := l.world.TagDefinition(inst.DefinitionID)
			if !ok {
				continue
			}
			for _, hook := range def.Hooks {
				if hook.Phase != HookPostAction || !hookMatchesVerb(hook, a.Verb) {
					continue
				}
				hook.Handler(ctx, inst.Params)
			}
		}
	}

	// 4. 响应
	if ctx.Blocked {
		if len(ctx.Events) > 0 {
			a.Resp <- ActionResult{Events: ctx.Events}
		} else {
			msgKey := "system.move.blocked"
			if ctx.BlockReason == "locked" {
				msgKey = "system.move.locked"
			}
			a.Resp <- ActionResult{
				Events: []presentation.Event{
					presentation.SystemMessageEvent{MessageKey: msgKey},
				},
			}
		}
		return
	}
	a.Resp <- ActionResult{Events: ctx.Events, NewRoom: ctx.NewRoom}

	// 5. 离开广播
	if ctx.OldRoom != "" && ctx.NewRoom != "" && ctx.OldRoom != ctx.NewRoom {
		l.SendToRoom(ctx.OldRoom, []presentation.Event{
			presentation.SystemMessageEvent{
				MessageKey: "system.player.left",
				Fields:     map[string]string{"direction": ctx.LeaveDir},
			},
		}, ctx.PlayerID)
		l.SendToRoom(ctx.NewRoom, []presentation.Event{
			presentation.SystemMessageEvent{
				MessageKey: "system.player.entered",
				Fields:     map[string]string{"direction": oppositeDirection(ctx.LeaveDir)},
			},
		}, ctx.PlayerID)
	}
}

// relevantItems 返回当前 action 上下文相关的物品。
// 钩子通过检查这些物品的 tag 实例来决定行为。
// 优先使用注册的 ItemResolver；如果没有注册，使用通用 fallback：
// 扫描当前房间+玩家背包中所有 hook.Verbs 匹配的物品。
func (l *Loop) relevantItems(ctx *AttemptContext) []Item {
	// 1. 查找已注册的 resolver
	if resolver, ok := l.itemResolvers[ctx.Verb]; ok {
		return resolver(l, ctx)
	}

	// 2. 通用 fallback：扫描可见物品中 hook.Verbs 匹配的物品
	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	return l.visibleItemsWithMatchingHooks(roomID, ctx.PlayerID, ctx.Verb)
}

// visibleItemsWithMatchingHooks 扫描当前房间+玩家背包，返回 tag hooks 匹配指定动词的物品。
func (l *Loop) visibleItemsWithMatchingHooks(roomID RoomID, playerID PlayerID, verb string) []Item {
	itemIDs := l.world.visibleItemIDs(roomID, playerID)
	var matched []Item
	for _, itemID := range itemIDs {
		item, ok := l.world.items[itemID]
		if !ok {
			continue
		}
		if itemHasHookForVerb(item, l.world, verb) {
			matched = append(matched, item)
		}
	}
	return matched
}

// itemHasHookForVerb 检查物品是否有 tag 定义包含匹配指定动词的 hook。
func itemHasHookForVerb(item Item, w *World, verb string) bool {
	for _, inst := range item.Tags {
		def, ok := w.TagDefinition(inst.DefinitionID)
		if !ok {
			continue
		}
		for _, hook := range def.Hooks {
			if hookMatchesVerb(hook, verb) {
				return true
			}
		}
	}
	return false
}

func handleEnterWorld(l *Loop, ctx *AttemptContext) {
	startRoom, ok := l.world.EnterWorld(ctx.PlayerID)
	if !ok {
		ctx.Blocked = true
		return
	}
	ctx.NewRoom = startRoom
}

func handleLeaveWorld(l *Loop, ctx *AttemptContext) {
	l.world.LeaveWorld(ctx.PlayerID)
}

func handleMove(l *Loop, ctx *AttemptContext) {
	nextRoom, ok := l.world.MovePlayer(ctx.PlayerID, ctx.Input)
	if !ok {
		ctx.Blocked = true
		return
	}
	obs, ok := l.world.Look(nextRoom)
	if !ok {
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, newRoomObservationEvent(obs))
	progEvents := l.applyProgression(ctx.PlayerID, progression.Trigger{
		Kind: progression.TriggerMovedRoom, RoomID: string(nextRoom),
	})
	ctx.Events = append(ctx.Events, progEvents...)
	ctx.NewRoom = nextRoom
	if ctx.OldRoom != "" && ctx.OldRoom != nextRoom {
		ctx.LeaveDir = ctx.Input
	}
}

func handleLook(l *Loop, ctx *AttemptContext) {
	roomID, ok := l.world.PlayerRoom(ctx.PlayerID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.player.not_found"})
		ctx.Blocked = true
		return
	}
	obs, ok := l.world.Look(roomID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.room.missing"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events,
		newRoomObservationEvent(obs),
		presentation.SystemMessageEvent{MessageKey: "system.look.observed"},
	)
	ctx.NewRoom = roomID
}

func handleGet(l *Loop, ctx *AttemptContext) {
	resolution := l.world.ResolveRoomItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.Input)
	if len(resolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)
		return
	}
	if !resolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	itemID, ok := l.world.GetItem(l.world.PlayerCurrentRoom(ctx.PlayerID), resolution.ItemID, ctx.PlayerID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events,
		presentation.SystemMessageEvent{
			MessageKey: "system.item.taken",
			Fields:     map[string]string{"item": string(itemID)},
		},
	)
	progEvents := l.applyProgression(ctx.PlayerID, progression.Trigger{
		Kind: progression.TriggerGotItem, ItemID: string(itemID),
	})
	ctx.Events = append(ctx.Events, progEvents...)
}

func handleDrop(l *Loop, ctx *AttemptContext) {
	resolution := l.world.ResolveInventoryItemPhrase(ctx.PlayerID, ctx.Input)
	if len(resolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)
		return
	}
	if !resolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"})
		ctx.Blocked = true
		return
	}
	itemID, ok := l.world.DropInventoryItem(l.world.PlayerCurrentRoom(ctx.PlayerID), resolution.ItemID, ctx.PlayerID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events,
		presentation.SystemMessageEvent{
			MessageKey: "system.item.dropped",
			Fields:     map[string]string{"item": string(itemID)},
		},
	)
}

func handleExamine(l *Loop, ctx *AttemptContext) {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, ctx.Input)
	if len(resolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)
		return
	}
	if !resolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	item, ok := l.world.ExamineItem(l.world.PlayerCurrentRoom(ctx.PlayerID), resolution.ItemID, ctx.PlayerID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, newItemObservationEvent(item))
	progEvents := l.applyProgression(ctx.PlayerID, progression.Trigger{
		Kind: progression.TriggerExaminedItem, ItemID: string(item.Item),
	})
	ctx.Events = append(ctx.Events, progEvents...)
}

func handleLookItem(l *Loop, ctx *AttemptContext) {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, ctx.Input)
	if len(resolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)
		return
	}
	if !resolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	item, ok := l.world.ExamineItem(l.world.PlayerCurrentRoom(ctx.PlayerID), resolution.ItemID, ctx.PlayerID)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, newItemObservationEvent(item))
}

func handleInventory(l *Loop, ctx *AttemptContext) {
	itemIDs := l.world.InventoryItemIDs(ctx.PlayerID)
	items := make([]string, 0, len(itemIDs))
	for _, id := range itemIDs {
		items = append(items, string(id))
	}
	ctx.Events = append(ctx.Events, presentation.InventoryEvent{Items: items})
}

func handleQuest(l *Loop, ctx *AttemptContext) {
	pid := string(ctx.PlayerID)
	status, ok := l.progression.Status(pid)
	if !ok {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.quest.none"})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, presentation.QuestStatusEvent{
		QuestID:    status.QuestID,
		QuestName:  status.QuestName,
		StageID:    status.StageID,
		StageText:  status.StageText,
		Conditions: status.Conditions,
		State:      string(status.State),
	})
}

func (l *Loop) applyProgression(playerID PlayerID, trigger progression.Trigger) []presentation.Event {
	status, advanced := l.progression.Apply(string(playerID), trigger)
	if !advanced {
		return nil
	}
	if status.State == progression.QuestStateRewardPending {
		resolvedStatus, resolved := l.progression.ResolveRewards(string(playerID))
		if resolved {
			status = resolvedStatus
		}
	}
	return []presentation.Event{
		presentation.SystemMessageEvent{
			MessageKey: "system.quest.progress",
			Fields: map[string]string{
				"quest_id": status.QuestID,
				"stage_id": status.StageID,
				"state":    string(status.State),
			},
		},
	}
}

func oppositeDirection(dir string) string {
	switch dir {
	case "north":
		return "south"
	case "south":
		return "north"
	case "east":
		return "west"
	case "west":
		return "east"
	case "northeast":
		return "southwest"
	case "northwest":
		return "southeast"
	case "southeast":
		return "northwest"
	case "southwest":
		return "northeast"
	case "up":
		return "down"
	case "down":
		return "up"
	default:
		return dir
	}
}
