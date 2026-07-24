package world

import (
	"log"
	"strings"
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
	l.RegisterVerb("open", handleOpen)
	l.RegisterVerb("close", handleClose)
	l.RegisterVerb("put", handlePut)
	l.RegisterVerb("get-from", handleGetFrom)
	l.RegisterVerb("light", handleLight)

	// 注册内置动词到注册表
	for name := range l.verbHandlers {
		if name == "" || name[0] == '_' {
			continue
		}
		l.verbRegistry[name] = VerbEntry{Name: name, Source: VerbSourceBuiltin}
	}

	l.RegisterItemResolver("move", resolveMoveEntity)
	l.RegisterItemResolver("get", resolveRoomEntityByPhrase)
	l.RegisterItemResolver("examine", resolveVisibleEntityByPhrase)
	l.RegisterItemResolver("look-item", resolveVisibleEntityByPhrase)
	l.RegisterItemResolver("drop", resolveInventoryEntityByPhrase)
	l.RegisterItemResolver("open", resolveVisibleEntityByPhrase)
	l.RegisterItemResolver("close", resolveVisibleEntityByPhrase)
	l.RegisterItemResolver("put", resolvePutEntities)
	l.RegisterItemResolver("get-from", resolveGetFromEntities)
	l.RegisterItemResolver("light", resolveVisibleEntityByPhrase)
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

// resolveMoveEntity 找出当前房间中匹配输入方向的那个出口实体。
func resolveMoveEntity(l *Loop, ctx *AttemptContext) []*Entity {
	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	for _, eid := range l.world.exitItemIDs(roomID) {
		ed := l.world.store.Exit(eid)
		if ed != nil && ed.Direction == ctx.Input {
			if ent := l.world.store.Get(eid); ent != nil {
				return []*Entity{ent}
			}
		}
		}
	}
	return nil
}

// resolveRoomEntityByPhrase 在当前房间中按短语匹配物品。
func resolveRoomEntityByPhrase(l *Loop, ctx *AttemptContext) []*Entity {
	resolution := l.world.ResolveRoomItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.Input)
	if resolution.Found {
		if ent := l.world.store.Get(resolution.ItemID); ent != nil {
			return []*Entity{ent}
		}
	}
	return nil
}

// resolveVisibleEntityByPhrase 在当前房间+玩家背包中按短语匹配 item 实体。
func resolveVisibleEntityByPhrase(l *Loop, ctx *AttemptContext) []*Entity {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, ctx.Input)
	if resolution.Found {
		if ent := l.world.store.Get(resolution.ItemID); ent != nil {
			return []*Entity{ent}
		}
	}
	return nil
}

// resolveInventoryEntityByPhrase 在玩家背包中按短语匹配 item 实体。
func resolveInventoryEntityByPhrase(l *Loop, ctx *AttemptContext) []*Entity {
	resolution := l.world.ResolveInventoryItemPhrase(ctx.PlayerID, ctx.Input)
	if resolution.Found {
		if ent := l.world.store.Get(resolution.ItemID); ent != nil {
			return []*Entity{ent}
		}
	}
	return nil
}

// resolvePutEntities 返回放物品动作涉及的所有 item 实体（物品+容器）
func resolvePutEntities(l *Loop, ctx *AttemptContext) []*Entity {
	itemPhrase, containerPhrase, ok := strings.Cut(ctx.Input, "|")
	if !ok || itemPhrase == "" || containerPhrase == "" {
		return nil
	}
	var result []*Entity
	if ents := resolveInventoryEntityByPhrase(l, &AttemptContext{
		PlayerID: ctx.PlayerID, Input: itemPhrase,
	}); ents != nil {
		result = append(result, ents...)
	}
	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	containerRes := l.world.ResolveVisibleItemPhrase(roomID, ctx.PlayerID, containerPhrase)
	if containerRes.Found {
		if ent := l.world.store.Get(containerRes.ItemID); ent != nil {
			result = append(result, ent)
		}
	}
	return result
}

// resolveGetFromEntities 返回从容器取物品动作涉及的所有 item 实体（容器+内部物品）
func resolveGetFromEntities(l *Loop, ctx *AttemptContext) []*Entity {
	itemPhrase, containerPhrase, ok := strings.Cut(ctx.Input, "|")
	if !ok || itemPhrase == "" || containerPhrase == "" {
		return nil
	}
	var result []*Entity
	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	containerRes := l.world.ResolveVisibleItemPhrase(roomID, ctx.PlayerID, containerPhrase)
	if !containerRes.Found {
		return nil
	}
	if ent := l.world.store.Get(containerRes.ItemID); ent != nil {
		result = append(result, ent)
	}
	for _, id := range l.world.ContainerContents(containerRes.ItemID) {
		ent := l.world.store.Get(id)
		if ent == nil {
			continue
		}
		if ent.matchesPhrase(id, itemPhrase) {
			result = append(result, ent)
			break
		}
	}
	return result
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

	entities := l.relevantEntities(ctx)
	for _, ent := range entities {
		for _, inst := range ent.Tags {
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

	if !ctx.Blocked {
		handler(l, ctx)
	}

	for _, ent := range l.relevantEntities(ctx) {
		for _, inst := range ent.Tags {
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

// relevantEntities 返回当前 action 上下文相关的实体。
// 钩子通过检查这些实体的 tag 实例来决定行为。
// 优先使用注册的 ItemResolver；如果没有注册，使用通用 fallback：
// 扫描当前房间+玩家背包中所有 hook.Verbs 匹配的实体。
func (l *Loop) relevantEntities(ctx *AttemptContext) []*Entity {
	if resolver, ok := l.itemResolvers[ctx.Verb]; ok {
		return resolver(l, ctx)
	}

	roomID := l.world.PlayerCurrentRoom(ctx.PlayerID)
	return l.visibleEntitiesWithMatchingHooks(roomID, ctx.PlayerID, ctx.Verb)
}

// visibleEntitiesWithMatchingHooks 扫描当前房间+玩家背包，返回 tag hooks 匹配指定动词的实体。
func (l *Loop) visibleEntitiesWithMatchingHooks(roomID, playerID EntityID, verb string) []*Entity {
	itemIDs := l.world.visibleItemIDs(roomID, playerID)
	var matched []*Entity
	for _, eid := range itemIDs {
		ent := l.world.store.Get(eid)
		if ent == nil {
			continue
		}
		if entityHasHookForVerb(ent, l.world, verb) {
			matched = append(matched, ent)
		}
	}
	return matched
}

// entityHasHookForVerb 检查实体是否有 tag 定义包含匹配指定动词的 hook。
func entityHasHookForVerb(entity *Entity, w *World, verb string) bool {
	for _, inst := range entity.Tags {
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
	tracked := l.world.TrackedQuest(ctx.PlayerID)
	if tracked == "" {
		for _, s := range l.progression.AllStatuses(string(ctx.PlayerID)) {
			if s.State == progression.QuestStateActive {
				l.world.SetTrackedQuest(ctx.PlayerID, s.QuestID)
				break
			}
		}
	}
	ctx.NewRoom = startRoom
}

func handleLeaveWorld(l *Loop, ctx *AttemptContext) {
	l.world.LeaveWorld(ctx.PlayerID)
}

func handleMove(l *Loop, ctx *AttemptContext) {
	if l.world.IsOverWeight(ctx.PlayerID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.move.overweight"})
		ctx.Blocked = true
		return
	}
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
	if obs.Dark && !l.world.RoomIsLit(nextRoom, ctx.PlayerID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.room.dark"})
	} else {
		ctx.Events = append(ctx.Events, newRoomObservationEvent(obs))
	}
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
	if !l.world.RoomIsLit(roomID, ctx.PlayerID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.room.dark"})
		ctx.NewRoom = roomID
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
	itemID, ok, volumeOk := l.world.GetItem(l.world.PlayerCurrentRoom(ctx.PlayerID), resolution.ItemID, ctx.PlayerID)
	if !ok {
		if !volumeOk {
			ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.inventory_full"})
			ctx.Blocked = true
			return
		}
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

func handleLight(l *Loop, ctx *AttemptContext) {
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
	itemID := resolution.ItemID
	ent := l.world.store.Get(itemID)
	if ent == nil {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	if !l.world.store.Tag(itemID, "tag.lightable") {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.light.not_lightable"})
		ctx.Blocked = true
		return
	}
	if !l.world.PlayerHasItem(ctx.PlayerID, itemID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.light.not_carried"})
		ctx.Blocked = true
		return
	}
	if l.world.IsItemLit(itemID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.light.already_lit"})
		ctx.Blocked = true
		return
	}
	l.world.LightItem(itemID)
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: "verb.light.default",
		Fields:     map[string]string{"item": ent.Name},
	})
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
	wtCur, wtMax := l.world.PlayerWeightRatio(ctx.PlayerID)
	volCur, volMax := l.world.PlayerVolumeRatio(ctx.PlayerID)
	ctx.Events = append(ctx.Events, presentation.InventoryEvent{
		Items:         items,
		WeightCurrent: wtCur,
		WeightMax:     wtMax,
		VolumeCurrent: volCur,
		VolumeMax:     volMax,
	})
}

func handleQuest(l *Loop, ctx *AttemptContext) {
	pid := string(ctx.PlayerID)
	if ctx.Input != "" {
		if strings.HasPrefix(ctx.Input, "accept ") {
			questID := strings.TrimSpace(strings.TrimPrefix(ctx.Input, "accept "))
			status, ok := l.progression.ActivateQuest(pid, questID)
			if !ok {
				ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.quest.not_unlocked"})
				ctx.Blocked = true
				return
			}
			l.world.SetTrackedQuest(ctx.PlayerID, questID)
			ctx.Events = append(ctx.Events, presentation.QuestStatusEvent{QuestID: status.QuestID, QuestName: status.QuestName, StageID: status.StageID, StageText: status.StageText, Conditions: status.Conditions, State: string(status.State)})
			return
		}
		status, ok := l.progression.Status(pid, ctx.Input)
		if !ok {
			ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.quest.none"})
			ctx.Blocked = true
			return
		}
		if status.State == progression.QuestStateHidden {
			ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.quest.none"})
			ctx.Blocked = true
			return
		}
		l.world.SetTrackedQuest(ctx.PlayerID, ctx.Input)
		ctx.Events = append(ctx.Events, presentation.QuestStatusEvent{
			QuestID:    status.QuestID,
			QuestName:  status.QuestName,
			StageID:    status.StageID,
			StageText:  status.StageText,
			Conditions: status.Conditions,
			State:      string(status.State),
		})
		return
	}
	allStatuses := l.progression.AllStatuses(pid)
	if len(allStatuses) == 0 {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.quest.none"})
		ctx.Blocked = true
		return
	}
	tracked := l.world.TrackedQuest(ctx.PlayerID)
	quests := make([]presentation.QuestStatusEvent, 0, len(allStatuses))
	for _, s := range allStatuses {
		if s.State == progression.QuestStateHidden {
			continue
		}
		quests = append(quests, presentation.QuestStatusEvent{
			QuestID:    s.QuestID,
			QuestName:  s.QuestName,
			StageID:    s.StageID,
			StageText:  s.StageText,
			Conditions: s.Conditions,
			State:      string(s.State),
		})
	}
	ctx.Events = append(ctx.Events, presentation.QuestListEvent{
		Quests:  quests,
		Tracked: tracked,
	})
}

func (l *Loop) applyProgression(playerID PlayerID, trigger progression.Trigger) []presentation.Event {
	statuses := l.progression.Apply(string(playerID), trigger)
	if len(statuses) == 0 {
		return nil
	}
	var events []presentation.Event
	for _, status := range statuses {
		if status.State == progression.QuestStateRewardPending {
			resolvedStatus, resolved := l.progression.ResolveRewards(string(playerID), status.QuestID)
			if resolved {
				status = resolvedStatus
			}
		}
		events = append(events,
			presentation.SystemMessageEvent{
				MessageKey: "system.quest.progress",
				Fields: map[string]string{
					"quest_id": status.QuestID,
					"stage_id": status.StageID,
					"state":    string(status.State),
				},
			},
			presentation.QuestStatusEvent{
				QuestID:    status.QuestID,
				QuestName:  status.QuestName,
				StageID:    status.StageID,
				StageText:  status.StageText,
				Conditions: status.Conditions,
				State:      string(status.State),
			},
		)
	}
	// Auto-advance tracked quest if current tracked quest completed
	tracked := l.world.TrackedQuest(playerID)
	if tracked != "" {
		s, ok := l.progression.Status(string(playerID), tracked)
		if ok && s.State == progression.QuestStateCompleted {
			// Find next active quest to track
			for _, qs := range l.progression.AllStatuses(string(playerID)) {
				if qs.State == progression.QuestStateActive {
					l.world.SetTrackedQuest(playerID, qs.QuestID)
					break
				}
			}
		}
	}
	return events
}

func handleOpen(l *Loop, ctx *AttemptContext) {
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
	itemName := l.world.ItemNameOr(resolution.ItemID)
	if !l.world.OpenContainer(resolution.ItemID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.container.cant_open",
			Fields:     map[string]string{"item": itemName},
		})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: "system.container.now_open",
		Fields:     map[string]string{"item": itemName},
	})
}

func handleClose(l *Loop, ctx *AttemptContext) {
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
	itemName := l.world.ItemNameOr(resolution.ItemID)
	if !l.world.CloseContainer(resolution.ItemID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.container.cant_close",
			Fields:     map[string]string{"item": itemName},
		})
		ctx.Blocked = true
		return
	}
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: "system.container.now_closed",
		Fields:     map[string]string{"item": itemName},
	})
}

func handlePut(l *Loop, ctx *AttemptContext) {
	itemPhrase, containerPhrase, ok := strings.Cut(ctx.Input, "|")
	if !ok || itemPhrase == "" || containerPhrase == "" {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.unknown_command"})
		ctx.Blocked = true
		return
	}
	itemResolution := l.world.ResolveInventoryItemPhrase(ctx.PlayerID, itemPhrase)
	if len(itemResolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, itemResolution.AmbiguousItemIDs)
		return
	}
	if !itemResolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"})
		ctx.Blocked = true
		return
	}
	containerResolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, containerPhrase)
	if len(containerResolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, containerResolution.AmbiguousItemIDs)
		return
	}
	if !containerResolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	err := l.world.PutItemInContainer(itemResolution.ItemID, containerResolution.ItemID, ctx.PlayerID)
	if err != nil {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.container.put_fail",
			Fields:     map[string]string{"error": err.Error()},
		})
		ctx.Blocked = true
		return
	}
	itemName := l.world.ItemNameOr(itemResolution.ItemID)
	containerName := l.world.ItemNameOr(containerResolution.ItemID)
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: "system.container.put_success",
		Fields: map[string]string{
			"item":      itemName,
			"container": containerName,
		},
	})
}

func handleGetFrom(l *Loop, ctx *AttemptContext) {
	itemPhrase, containerPhrase, ok := strings.Cut(ctx.Input, "|")
	if !ok || itemPhrase == "" || containerPhrase == "" {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.unknown_command"})
		ctx.Blocked = true
		return
	}
	containerResolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(ctx.PlayerID), ctx.PlayerID, containerPhrase)
	if len(containerResolution.AmbiguousItemIDs) > 0 {
		ctx.Events = ambiguousItemsEvent(l.world, containerResolution.AmbiguousItemIDs)
		return
	}
	if !containerResolution.Found {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		ctx.Blocked = true
		return
	}
	containerName := l.world.ItemNameOr(containerResolution.ItemID)
	if !l.world.ContainerIsOpen(containerResolution.ItemID) {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.container.closed",
			Fields:     map[string]string{"container": containerName},
		})
		ctx.Blocked = true
		return
	}
	contents := l.world.ContainerContents(containerResolution.ItemID)
	var matches []EntityID
	for _, id := range contents {
		ent := l.world.store.Get(id)
		if ent == nil {
			continue
		}
		if ent.matchesPhrase(id, itemPhrase) {
			matches = append(matches, id)
		}
	}
	if len(matches) > 1 {
		ctx.Events = ambiguousItemsEvent(l.world, matches)
		return
	}
	if len(matches) == 0 {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.item.not_in_container",
			Fields:     map[string]string{"container": containerName},
		})
		ctx.Blocked = true
		return
	}
	err := l.world.GetItemFromContainer(containerResolution.ItemID, matches[0], ctx.PlayerID)
	if err != nil {
		ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
			MessageKey: "system.container.get_fail",
			Fields:     map[string]string{"error": err.Error()},
		})
		ctx.Blocked = true
		return
	}
	itemName := l.world.ItemNameOr(matches[0])
	ctx.Events = append(ctx.Events, presentation.SystemMessageEvent{
		MessageKey: "system.container.get_success",
		Fields: map[string]string{
			"item":      itemName,
			"container": containerName,
		},
	})
	progEvents := l.applyProgression(ctx.PlayerID, progression.Trigger{
		Kind: progression.TriggerGotItem, ItemID: string(matches[0]),
	})
	ctx.Events = append(ctx.Events, progEvents...)
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
