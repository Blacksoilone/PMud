package presentation

import "testing"

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
		Room:           "room.tutorial.start",
		NameKey:        "room.tutorial.start.name",
		DescriptionKey: "room.tutorial.start.description",
		Name:           "练习场入口",
		Description:    "这里是练习场的入口。北边传来木剑碰撞的声音。",
		Exits:          []string{"north"},
		Neighbors:      map[string]string{"north": "room.tutorial.yard"},
		Items:          []string{"item.tutorial.old_lantern"},
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\tneighbors=north=room.tutorial.yard\titems=item.tutorial.old_lantern\n"
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
	want := "event=inventory\titems=item.tutorial.old_lantern,item.tutorial.practice_sword\n"
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
	want := "event=inventory\titems=\n"
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
