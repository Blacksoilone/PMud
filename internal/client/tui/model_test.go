package tui

import (
	"testing"

	"PMud/internal/protocol"
)

func TestNewModelStartsEmpty(t *testing.T) {
	model := NewModel(2)

	if len(model.Events) != 0 {
		t.Fatalf("Events length = %d, want 0", len(model.Events))
	}
	if model.Input != "" {
		t.Fatalf("Input = %q, want empty", model.Input)
	}
	if model.HistoryLimit != 2 {
		t.Fatalf("HistoryLimit = %d, want 2", model.HistoryLimit)
	}
	if len(model.Regions.Log) != 0 {
		t.Fatalf("Regions.Log length = %d, want 0", len(model.Regions.Log))
	}
	if model.Regions.Room.Room != "" {
		t.Fatalf("Regions.Room.Room = %q, want empty", model.Regions.Room.Room)
	}
	if model.Regions.Inventory.Items != "" {
		t.Fatalf("Regions.Inventory.Items = %q, want empty", model.Regions.Inventory.Items)
	}
	if model.Regions.Quest.QuestID != "" {
		t.Fatalf("Regions.Quest.QuestID = %q, want empty", model.Regions.Quest.QuestID)
	}
	if model.Regions.Item.Item != "" {
		t.Fatalf("Regions.Item.Item = %q, want empty", model.Regions.Item.Item)
	}
	if model.Regions.QuestNotice.MessageKey != "" {
		t.Fatalf("Regions.QuestNotice.MessageKey = %q, want empty", model.Regions.QuestNotice.MessageKey)
	}
}

func TestNewModelClampsHistoryLimit(t *testing.T) {
	model := NewModel(0)

	if model.HistoryLimit != 1 {
		t.Fatalf("HistoryLimit = %d, want 1", model.HistoryLimit)
	}
}

func TestApplyEventAppendsEvents(t *testing.T) {
	model := NewModel(2)
	roomEvent := protocol.Event{Name: "room", Fields: map[string]string{"room": "room.tutorial.start"}}
	inventoryEvent := protocol.Event{Name: "inventory"}

	model = ApplyEvent(model, roomEvent)
	model = ApplyEvent(model, inventoryEvent)

	if len(model.Events) != 2 {
		t.Fatalf("Events length = %d, want 2", len(model.Events))
	}
	if model.Events[0].Name != "room" || model.Events[1].Name != "inventory" {
		t.Fatalf("Events = %#v, want room then inventory", model.Events)
	}
}

func TestApplyEventDropsOldestEvent(t *testing.T) {
	model := NewModel(2)

	model = ApplyEvent(model, protocol.Event{Name: "first"})
	model = ApplyEvent(model, protocol.Event{Name: "second"})
	model = ApplyEvent(model, protocol.Event{Name: "third"})

	if len(model.Events) != 2 {
		t.Fatalf("Events length = %d, want 2", len(model.Events))
	}
	if model.Events[0].Name != "second" || model.Events[1].Name != "third" {
		t.Fatalf("Events = %#v, want second then third", model.Events)
	}
	if len(model.Regions.Log) != 2 {
		t.Fatalf("Regions.Log length = %d, want 2", len(model.Regions.Log))
	}
	if model.Regions.Log[0].Name != "second" || model.Regions.Log[1].Name != "third" {
		t.Fatalf("Regions.Log = %#v, want second then third", model.Regions.Log)
	}
}

func TestApplyEventUpdatesRoomRegion(t *testing.T) {
	model := NewModel(3)

	model = ApplyEvent(model, protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            "room.tutorial.start",
			"name_key":        "room.tutorial.start.name",
			"description_key": "room.tutorial.start.description",
			"exits":           "north",
			"items":           "item.tutorial.old_lantern",
		},
	})

	if model.Regions.Room.Room != "room.tutorial.start" {
		t.Fatalf("Regions.Room.Room = %q", model.Regions.Room.Room)
	}
	if model.Regions.Room.NameKey != "room.tutorial.start.name" {
		t.Fatalf("Regions.Room.NameKey = %q", model.Regions.Room.NameKey)
	}
	if model.Regions.Room.DescriptionKey != "room.tutorial.start.description" {
		t.Fatalf("Regions.Room.DescriptionKey = %q", model.Regions.Room.DescriptionKey)
	}
	if model.Regions.Room.Exits != "north" {
		t.Fatalf("Regions.Room.Exits = %q", model.Regions.Room.Exits)
	}
	if model.Regions.Room.Items != "item.tutorial.old_lantern" {
		t.Fatalf("Regions.Room.Items = %q", model.Regions.Room.Items)
	}
}

func TestApplyEventUpdatesInventoryRegion(t *testing.T) {
	model := NewModel(3)

	model = ApplyEvent(model, protocol.Event{
		Name:   "inventory",
		Fields: map[string]string{"items": "item.tutorial.old_lantern,item.tutorial.practice_sword"},
	})

	if model.Regions.Inventory.Items != "item.tutorial.old_lantern,item.tutorial.practice_sword" {
		t.Fatalf("Regions.Inventory.Items = %q", model.Regions.Inventory.Items)
	}
}

func TestApplyEventUpdatesQuestRegion(t *testing.T) {
	model := NewModel(3)

	model = ApplyEvent(model, protocol.Event{
		Name: "quest",
		Fields: map[string]string{
			"quest_id":   "quest.tutorial.first_steps",
			"quest_name": "初入练习场",
			"stage_id":   "quest.tutorial.first_steps.stage.enter_yard",
			"stage_text": "走进院子",
			"conditions": "向北移动",
			"state":      "active",
		},
	})

	if model.Regions.Quest.QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("Regions.Quest.QuestID = %q", model.Regions.Quest.QuestID)
	}
	if model.Regions.Quest.QuestName != "初入练习场" {
		t.Fatalf("Regions.Quest.QuestName = %q", model.Regions.Quest.QuestName)
	}
	if model.Regions.Quest.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("Regions.Quest.StageID = %q", model.Regions.Quest.StageID)
	}
	if model.Regions.Quest.StageText != "走进院子" {
		t.Fatalf("Regions.Quest.StageText = %q", model.Regions.Quest.StageText)
	}
	if model.Regions.Quest.Conditions != "向北移动" {
		t.Fatalf("Regions.Quest.Conditions = %q", model.Regions.Quest.Conditions)
	}
	if model.Regions.Quest.State != "active" {
		t.Fatalf("Regions.Quest.State = %q", model.Regions.Quest.State)
	}
}

func TestApplyEventUpdatesItemRegion(t *testing.T) {
	model := NewModel(3)

	model = ApplyEvent(model, protocol.Event{
		Name: "item",
		Fields: map[string]string{
			"item":            "item.tutorial.old_lantern",
			"name_key":        "item.tutorial.old_lantern.name",
			"description_key": "item.tutorial.old_lantern.description",
		},
	})

	if model.Regions.Item.Item != "item.tutorial.old_lantern" {
		t.Fatalf("Regions.Item.Item = %q", model.Regions.Item.Item)
	}
	if model.Regions.Item.NameKey != "item.tutorial.old_lantern.name" {
		t.Fatalf("Regions.Item.NameKey = %q", model.Regions.Item.NameKey)
	}
	if model.Regions.Item.DescriptionKey != "item.tutorial.old_lantern.description" {
		t.Fatalf("Regions.Item.DescriptionKey = %q", model.Regions.Item.DescriptionKey)
	}
}

func TestApplyEventQuestProgressUpdatesNoticeWithoutOverwritingQuestRegion(t *testing.T) {
	model := NewModel(3)
	model = ApplyEvent(model, protocol.Event{
		Name: "quest",
		Fields: map[string]string{
			"quest_id":   "quest.tutorial.first_steps",
			"quest_name": "初入练习场",
			"stage_id":   "quest.tutorial.first_steps.stage.get_lantern",
			"stage_text": "拿起旧油灯",
			"conditions": "拿起旧油灯",
			"state":      "active",
		},
	})

	model = ApplyEvent(model, protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.quest.progress",
			"quest_id":    "quest.tutorial.first_steps",
			"stage_id":    "quest.tutorial.first_steps.stage.enter_yard",
			"state":       "active",
		},
	})

	if model.Regions.Quest.StageText != "拿起旧油灯" {
		t.Fatalf("Regions.Quest.StageText = %q, want full quest details preserved", model.Regions.Quest.StageText)
	}
	if model.Regions.QuestNotice.MessageKey != "system.quest.progress" {
		t.Fatalf("Regions.QuestNotice.MessageKey = %q", model.Regions.QuestNotice.MessageKey)
	}
	if model.Regions.QuestNotice.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("Regions.QuestNotice.StageID = %q", model.Regions.QuestNotice.StageID)
	}
}

func TestApplyEventHistoryLimitDoesNotTrimCurrentRegionSnapshots(t *testing.T) {
	model := NewModel(1)
	model = ApplyEvent(model, protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            "room.tutorial.start",
			"name_key":        "room.tutorial.start.name",
			"description_key": "room.tutorial.start.description",
		},
	})
	model = ApplyEvent(model, protocol.Event{
		Name:   "inventory",
		Fields: map[string]string{"items": "item.tutorial.old_lantern"},
	})

	if len(model.Events) != 1 {
		t.Fatalf("Events length = %d, want 1", len(model.Events))
	}
	if len(model.Regions.Log) != 1 {
		t.Fatalf("Regions.Log length = %d, want 1", len(model.Regions.Log))
	}
	if model.Regions.Room.Room != "room.tutorial.start" {
		t.Fatalf("Regions.Room snapshot was trimmed: %#v", model.Regions.Room)
	}
	if model.Regions.Inventory.Items != "item.tutorial.old_lantern" {
		t.Fatalf("Regions.Inventory.Items = %q", model.Regions.Inventory.Items)
	}
}

func TestApplyEventDoesNotMutatePreviousModel(t *testing.T) {
	model := NewModel(2)
	next := ApplyEvent(model, protocol.Event{Name: "room"})

	if len(model.Events) != 0 {
		t.Fatalf("original Events length = %d, want 0", len(model.Events))
	}
	if len(next.Events) != 1 {
		t.Fatalf("next Events length = %d, want 1", len(next.Events))
	}
}
