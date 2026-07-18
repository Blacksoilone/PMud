package world

import (
	"sync"

	"PMud/internal/presentation"
	"PMud/internal/progression"
)

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
}

func NewLoop(w *World) *Loop {
	return &Loop{
		world:       w,
		progression: progression.NewEngine(w.ProgressionDefinitions()),
		actions:     make(chan Action, 64),
		sessions:    make(map[PlayerID]chan<- []presentation.Event),
	}
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
	l.Submit(Action{PlayerID: playerID, Verb: "_leave_world"})
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
		l.handle(a)
	}
}

func (l *Loop) handle(a Action) {
	switch a.Verb {
	case "_enter_world":
		startRoom, ok := l.world.EnterWorld(a.PlayerID)
		if !ok {
			a.Resp <- ActionResult{}
			return
		}
		a.Resp <- ActionResult{NewRoom: startRoom}

	case "_leave_world":
		l.world.LeaveWorld(a.PlayerID)

	default:
		l.handleAction(a)
	}
}

func (l *Loop) handleAction(a Action) {
	switch a.Verb {
	case "move":
		l.handleMove(a)
	case "look":
		l.handleLook(a)
	case "get":
		l.handleGet(a)
	case "drop":
		l.handleDrop(a)
	case "examine":
		l.handleExamine(a)
	case "look-item":
		l.handleLookItem(a)
	case "inventory":
		l.handleInventory(a)
	case "quest":
		l.handleQuest(a)
	default:
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.unknown_command"},
			},
		}
	}
}

func (l *Loop) handleMove(a Action) {
	oldRoom := l.world.PlayerCurrentRoom(a.PlayerID)
	nextRoom, ok, reason := l.world.MovePlayer(a.PlayerID, a.Target)
	if !ok {
		msgKey := "system.move.blocked"
		if reason == "locked" {
			msgKey = "system.move.locked"
		}
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: msgKey},
			},
		}
		return
	}
	obs, ok := l.world.Look(nextRoom)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.room.missing"},
			},
		}
		return
	}
	events := []presentation.Event{newRoomObservationEvent(obs)}
	progEvents := l.applyProgression(a.PlayerID, progression.Trigger{
		Kind: progression.TriggerMovedRoom, RoomID: string(nextRoom),
	})
	events = append(events, progEvents...)
	a.Resp <- ActionResult{Events: events, NewRoom: nextRoom}

	if oldRoom != nextRoom {
		l.SendToRoom(oldRoom, []presentation.Event{
			presentation.SystemMessageEvent{
				MessageKey: "system.player.left",
				Fields:     map[string]string{"direction": a.Target},
			},
		}, a.PlayerID)
	}
	l.SendToRoom(nextRoom, []presentation.Event{
		presentation.SystemMessageEvent{
			MessageKey: "system.player.entered",
			Fields:     map[string]string{"direction": oppositeDirection(a.Target)},
		},
	}, a.PlayerID)
}

func (l *Loop) handleLook(a Action) {
	roomID, ok := l.world.PlayerRoom(a.PlayerID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.player.not_found"},
			},
		}
		return
	}
	obs, ok := l.world.Look(roomID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.room.missing"},
			},
		}
		return
	}
	a.Resp <- ActionResult{
		Events: []presentation.Event{
			newRoomObservationEvent(obs),
			presentation.SystemMessageEvent{MessageKey: "system.look.observed"},
		},
		NewRoom: roomID,
	}
}

func (l *Loop) handleGet(a Action) {
	resolution := l.world.ResolveRoomItemPhrase(l.world.PlayerCurrentRoom(a.PlayerID), a.Target)
	if len(resolution.AmbiguousItemIDs) > 0 {
		a.Resp <- ActionResult{Events: ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)}
		return
	}
	if !resolution.Found {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	itemID, ok := l.world.GetItem(l.world.PlayerCurrentRoom(a.PlayerID), resolution.ItemID, a.PlayerID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	events := []presentation.Event{
		presentation.SystemMessageEvent{
			MessageKey: "system.item.taken",
			Fields:     map[string]string{"item": string(itemID)},
		},
	}
	progEvents := l.applyProgression(a.PlayerID, progression.Trigger{
		Kind: progression.TriggerGotItem, ItemID: string(itemID),
	})
	events = append(events, progEvents...)
	a.Resp <- ActionResult{Events: events}
}

func (l *Loop) handleDrop(a Action) {
	resolution := l.world.ResolveInventoryItemPhrase(a.PlayerID, a.Target)
	if len(resolution.AmbiguousItemIDs) > 0 {
		a.Resp <- ActionResult{Events: ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)}
		return
	}
	if !resolution.Found {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"},
			},
		}
		return
	}
	itemID, ok := l.world.DropInventoryItem(l.world.PlayerCurrentRoom(a.PlayerID), resolution.ItemID, a.PlayerID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"},
			},
		}
		return
	}
	a.Resp <- ActionResult{
		Events: []presentation.Event{
			presentation.SystemMessageEvent{
				MessageKey: "system.item.dropped",
				Fields:     map[string]string{"item": string(itemID)},
			},
		},
	}
}

func (l *Loop) handleExamine(a Action) {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(a.PlayerID), a.PlayerID, a.Target)
	if len(resolution.AmbiguousItemIDs) > 0 {
		a.Resp <- ActionResult{Events: ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)}
		return
	}
	if !resolution.Found {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	item, ok := l.world.ExamineItem(l.world.PlayerCurrentRoom(a.PlayerID), resolution.ItemID, a.PlayerID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	events := []presentation.Event{newItemObservationEvent(item)}
	progEvents := l.applyProgression(a.PlayerID, progression.Trigger{
		Kind: progression.TriggerExaminedItem, ItemID: string(item.Item),
	})
	events = append(events, progEvents...)
	a.Resp <- ActionResult{Events: events}
}

func (l *Loop) handleLookItem(a Action) {
	resolution := l.world.ResolveVisibleItemPhrase(l.world.PlayerCurrentRoom(a.PlayerID), a.PlayerID, a.Target)
	if len(resolution.AmbiguousItemIDs) > 0 {
		a.Resp <- ActionResult{Events: ambiguousItemsEvent(l.world, resolution.AmbiguousItemIDs)}
		return
	}
	if !resolution.Found {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	item, ok := l.world.ExamineItem(l.world.PlayerCurrentRoom(a.PlayerID), resolution.ItemID, a.PlayerID)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.item.not_here"},
			},
		}
		return
	}
	a.Resp <- ActionResult{Events: []presentation.Event{newItemObservationEvent(item)}}
}

func (l *Loop) handleInventory(a Action) {
	itemIDs := l.world.InventoryItemIDs(a.PlayerID)
	items := make([]string, 0, len(itemIDs))
	for _, id := range itemIDs {
		items = append(items, string(id))
	}
	a.Resp <- ActionResult{
		Events: []presentation.Event{
			presentation.InventoryEvent{Items: items},
		},
	}
}

func (l *Loop) handleQuest(a Action) {
	pid := string(a.PlayerID)
	status, ok := l.progression.Status(pid)
	if !ok {
		a.Resp <- ActionResult{
			Events: []presentation.Event{
				presentation.SystemMessageEvent{MessageKey: "system.quest.none"},
			},
		}
		return
	}
	a.Resp <- ActionResult{
		Events: []presentation.Event{
			presentation.QuestStatusEvent{
				QuestID:    status.QuestID,
				QuestName:  status.QuestName,
				StageID:    status.StageID,
				StageText:  status.StageText,
				Conditions: status.Conditions,
				State:      string(status.State),
			},
		},
	}
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
