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
