package render

import (
	"PMud/internal/protocol"
	"testing"
)

func TestRenderBlock_roomEventUsesClientCatalog(t *testing.T) {
	catalog := testCatalog()
	event := protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            "room.tutorial.hall",
			"name_key":        "room.tutorial.hall.name",
			"description_key": "room.tutorial.hall.description",
			"exits":           "north,east,portal",
		},
	}

	got := RenderBlock(event, catalog).String()
	want := "教学大厅\n大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。\n出口: north, east, portal\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRenderBlock_canBeBoxedByLayout(t *testing.T) {
	catalog := testCatalog()
	event := protocol.Event{
		Name: "inventory",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern",
		},
	}

	got := RenderBlock(event, catalog).Lines
	want := []string{"你带着: 旧油灯（old lantern）"}
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d", len(want), len(got))
	}
	if got[0] != want[0] {
		t.Fatalf("expected %q, got %q", want[0], got[0])
	}
}
