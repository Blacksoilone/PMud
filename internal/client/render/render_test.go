package render

import (
	"PMud/internal/content"
	"PMud/internal/protocol"
	"testing"
)

func TestRender_roomEventUsesClientCatalog(t *testing.T) {
	// Given
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

	// When
	got := Render(event, catalog)

	// Then
	want := "练习场入口\n这里是练习场的入口。北边传来木剑碰撞的声音。\n出口: north\n你看到: 旧油灯\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_inventoryEventUsesItemResourceKeys(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "inventory",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern,item.tutorial.practice_sword",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "你带着: 旧油灯, 练习木剑\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_systemEventUsesMessageKey(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.help",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "可用命令: look, go <direction>, get <item>, drop <item>, inventory, help\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_missingResourceFallsBackToKey(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"name_key":        "room.unknown.name",
			"description_key": "room.unknown.description",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "room.unknown.name\nroom.unknown.description\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func testCatalog() content.ClientCatalog {
	return content.ClientCatalog{
		ItemNames: map[content.ItemID]content.TextKey{
			"item.tutorial.old_lantern":    "item.tutorial.old_lantern.name",
			"item.tutorial.practice_sword": "item.tutorial.practice_sword.name",
		},
		Text: map[content.TextKey]string{
			"room.tutorial.start.name":          "练习场入口",
			"room.tutorial.start.description":   "这里是练习场的入口。北边传来木剑碰撞的声音。",
			"item.tutorial.old_lantern.name":    "旧油灯",
			"item.tutorial.practice_sword.name": "练习木剑",
			"system.help":                       "可用命令: look, go <direction>, get <item>, drop <item>, inventory, help",
		},
	}
}
