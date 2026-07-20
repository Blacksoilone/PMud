package presentation

import (
	"strings"
	"testing"
)

func TestTextRenderer_RenderSystemMessageEvent_asStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := SystemMessageEvent{MessageKey: "system.empty_input"}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=system\tmessage_key=system.empty_input\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderSystemMessageEvent_withFieldsAsStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := SystemMessageEvent{
		MessageKey: "system.item.taken",
		Fields: map[string]string{
			"item": "item.tutorial.old_lantern",
		},
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=system\tmessage_key=system.item.taken\titem=item.tutorial.old_lantern\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderRoomObservationEvent_asStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := RoomObservationEvent{
		Room:           "room.tutorial.hall",
		NameKey:        "room.tutorial.hall.name",
		DescriptionKey: "room.tutorial.hall.description",
		Name:           "教学大厅",
		Description:    "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。",
		Exits:          []string{"north", "east", "portal"},
		Neighbors:      map[string]string{"north": "room.tutorial.item_yard", "east": "room.tutorial.lock_hall"},
		Items:          []string{},
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=room\troom=room.tutorial.hall\tname_key=room.tutorial.hall.name\tdescription_key=room.tutorial.hall.description\texits=north,east,portal\tneighbors=east=room.tutorial.lock_hall,north=room.tutorial.item_yard\titems=\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderInventoryEvent_withItemsAsStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := InventoryEvent{Items: []string{"item.tutorial.old_lantern", "item.tutorial.practice_sword"}}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=inventory\titems=item.tutorial.old_lantern,item.tutorial.practice_sword\tweight=0/0\tvolume=0/0\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderInventoryEvent_withoutItemsAsStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := InventoryEvent{Items: nil}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=inventory\titems=\tweight=0/0\tvolume=0/0\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderItemObservationEvent_asStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := ItemObservationEvent{
		Item:           "item.tutorial.old_lantern",
		NameKey:        "item.tutorial.old_lantern.name",
		DescriptionKey: "item.tutorial.old_lantern.description",
		Name:           "旧油灯",
		Description:    "灯罩上蒙着一层灰，里面还剩一点灯油。",
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=item\titem=item.tutorial.old_lantern\tname_key=item.tutorial.old_lantern.name\tdescription_key=item.tutorial.old_lantern.description\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderEscapesFieldSeparators(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := SystemMessageEvent{
		MessageKey: "system.escaped",
		Fields: map[string]string{
			"input": "第一行\n第二行\t反斜杠\\",
		},
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=system\tmessage_key=system.escaped\tinput=第一行\\n第二行\\t反斜杠\\\\\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestTextRenderer_RenderQuestListEscapesStructuredText(t *testing.T) {
	line := (TextRenderer{}).Render(QuestListEvent{Quests: []QuestStatusEvent{{
		QuestID:   "quest.one",
		QuestName: "任务,一|特别",
		StageText: "阶段\n说明",
		State:     "active",
	}}})
	if !strings.Contains(line, `items=[{"id":"quest.one","name":"任务,一|特别","stage":"阶段\\n说明","state":"active"}]`) {
		t.Fatalf("quest list line = %q", line)
	}
}
