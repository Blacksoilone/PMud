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
			"room":            "room.tutorial.start",
			"name_key":        "room.tutorial.start.name",
			"description_key": "room.tutorial.start.description",
			"exits":           "north",
			"items":           "item.tutorial.old_lantern",
		},
	}

	got := RenderBlock(event, catalog).String()
	want := "练习场入口\n这里是练习场的入口。北边传来木剑碰撞的声音。\n出口: north\n你看到: 旧油灯\n"
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
	want := []string{"你带着: 旧油灯"}
	if len(got) != len(want) {
		t.Fatalf("expected %d lines, got %d", len(want), len(got))
	}
	if got[0] != want[0] {
		t.Fatalf("expected %q, got %q", want[0], got[0])
	}
}
