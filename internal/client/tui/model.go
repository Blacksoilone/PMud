package tui

import "PMud/internal/protocol"

type Model struct {
	Events       []protocol.Event
	Regions      RegionState
	Input        string
	HistoryLimit int
}

type RegionState struct {
	Log         []protocol.Event
	Room        RoomRegion
	Inventory   InventoryRegion
	Quest       QuestRegion
	Item        ItemRegion
	QuestNotice QuestNoticeRegion
}

type RoomRegion struct {
	Room           string
	NameKey        string
	DescriptionKey string
	Exits          string
	Items          string
}

type InventoryRegion struct {
	Items string
}

type QuestRegion struct {
	QuestID    string
	QuestName  string
	StageID    string
	StageText  string
	Conditions string
	State      string
}

type ItemRegion struct {
	Item           string
	NameKey        string
	DescriptionKey string
}

type QuestNoticeRegion struct {
	MessageKey string
	QuestID    string
	StageID    string
	State      string
}

type Command struct {
	Line      string
	Submitted bool
}

func NewModel(historyLimit int) Model {
	return Model{HistoryLimit: normalizeHistoryLimit(historyLimit)}
}

func ApplyEvent(model Model, event protocol.Event) Model {
	limit := normalizeHistoryLimit(model.HistoryLimit)
	events := append([]protocol.Event(nil), model.Events...)
	events = append(events, event)
	if len(events) > limit {
		events = events[len(events)-limit:]
	}
	model.Events = events
	model.Regions.Log = events
	model.Regions = applyRegionEvent(model.Regions, event)
	model.HistoryLimit = limit
	return model
}

func applyRegionEvent(regions RegionState, event protocol.Event) RegionState {
	switch event.Name {
	case "room":
		regions.Room = RoomRegion{
			Room:           event.Fields["room"],
			NameKey:        event.Fields["name_key"],
			DescriptionKey: event.Fields["description_key"],
			Exits:          event.Fields["exits"],
			Items:          event.Fields["items"],
		}
	case "inventory":
		regions.Inventory = InventoryRegion{Items: event.Fields["items"]}
	case "quest":
		regions.Quest = QuestRegion{
			QuestID:    event.Fields["quest_id"],
			QuestName:  event.Fields["quest_name"],
			StageID:    event.Fields["stage_id"],
			StageText:  event.Fields["stage_text"],
			Conditions: event.Fields["conditions"],
			State:      event.Fields["state"],
		}
	case "item":
		regions.Item = ItemRegion{
			Item:           event.Fields["item"],
			NameKey:        event.Fields["name_key"],
			DescriptionKey: event.Fields["description_key"],
		}
	case "system":
		if event.Fields["message_key"] == "system.quest.progress" {
			regions.QuestNotice = QuestNoticeRegion{
				MessageKey: event.Fields["message_key"],
				QuestID:    event.Fields["quest_id"],
				StageID:    event.Fields["stage_id"],
				State:      event.Fields["state"],
			}
		}
	}
	return regions
}

func normalizeHistoryLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	return limit
}
