package render

import (
	"testing"

	"PMud/internal/content"
	"PMud/internal/protocol"
)

func TestRender_roomEventUsesClientCatalog(t *testing.T) {
	// Given
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

	// When
	got := Render(event, catalog)

	// Then
	want := "教学大厅\n大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。\n出口: north, east, portal\n"
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
	want := "你带着: 旧油灯（old lantern）, 练习木剑（practice sword）\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_itemEventUsesClientCatalog(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "item",
		Fields: map[string]string{
			"item":            "item.tutorial.old_lantern",
			"name_key":        "item.tutorial.old_lantern.name",
			"description_key": "item.tutorial.old_lantern.description",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "旧油灯（old lantern）\n灯罩上蒙着一层灰，里面还剩一点灯油。\n"
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
	want := "可用命令: look, go <direction>, get <item>, drop <item>, inventory, quest, help\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_systemEventReplacesItemParameter(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.item.taken",
			"item":        "item.tutorial.old_lantern",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "你拿起了旧油灯（old lantern）。\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_systemEventReplacesInputParameter(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "system",
		Fields: map[string]string{
			"message_key": "system.unknown_command",
			"input":       "dance",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "未知命令: dance\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRender_questStatusEvent(t *testing.T) {
	// Given
	catalog := testCatalog()
	event := protocol.Event{
		Name: "quest",
		Fields: map[string]string{
			"quest_id":   "quest.tutorial.first_steps",
			"quest_name": "教程任务",
			"stage_id":   "quest.tutorial.first_steps.stage.get_lantern",
			"stage_text": "拿起旧油灯。",
			"conditions": "获取旧油灯",
			"state":      "active",
		},
	}

	// When
	got := Render(event, catalog)

	// Then
	want := "任务: 教程任务\n阶段: 拿起旧油灯。\n状态: active\n条件:\n- 获取旧油灯\n"
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
		ItemDisplayNames: map[content.ItemID]content.TextKey{
			"item.tutorial.old_lantern":    "item.tutorial.old_lantern.name",
			"item.tutorial.practice_sword": "item.tutorial.practice_sword.name",
		},
		ItemInnerNames: map[content.ItemID]content.TextKey{
			"item.tutorial.old_lantern":    "item.tutorial.old_lantern.inner_name",
			"item.tutorial.practice_sword": "item.tutorial.practice_sword.inner_name",
		},
		ItemDescriptions: map[content.ItemID]content.TextKey{
			"item.tutorial.old_lantern": "item.tutorial.old_lantern.description",
		},
		Text: map[content.TextKey]string{
			"room.tutorial.hall.name":                "教学大厅",
			"room.tutorial.hall.description":         "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。",
			"item.tutorial.old_lantern.name":          "旧油灯",
			"item.tutorial.old_lantern.inner_name":    "old lantern",
			"item.tutorial.old_lantern.description":   "灯罩上蒙着一层灰，里面还剩一点灯油。",
			"item.tutorial.practice_sword.name":       "练习木剑",
			"item.tutorial.practice_sword.inner_name": "practice sword",
			"system.help":            "可用命令: look, go <direction>, get <item>, drop <item>, inventory, quest, help",
			"system.item.taken":      "你拿起了{item}。",
			"system.unknown_command": "未知命令: {input}",
		},
	}
}
