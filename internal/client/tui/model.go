package tui

import "PMud/internal/protocol"

type Model struct {
	Events           []protocol.Event
	Regions          RegionState
	Input            string
	Editor           LineEditor
	HistoryLimit     int
	ExitConfirmation bool
	Popup            PopupState
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
	Neighbors      string
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
	Line          string
	Submitted     bool
	ExitRequested bool
}

type PopupContent struct {
	Kind  PopupKind
	Title string
	Lines []string
}

type PopupKind int

const (
	PopupNone PopupKind = iota
	PopupHelp
	PopupInventory
)

type PopupState struct {
	Active       bool
	Content      PopupContent
	ScrollOffset int
}

func NewModel(historyLimit int) Model {
	limit := normalizeHistoryLimit(historyLimit)
	return Model{Editor: NewLineEditor(defaultHistoryLimit), HistoryLimit: limit}
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
			Neighbors:      event.Fields["neighbors"],
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
