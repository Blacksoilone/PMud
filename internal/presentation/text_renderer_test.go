package presentation

import "testing"

func TestTextRenderer_RenderSystemMessageEvent_asStructuredLine(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := SystemMessageEvent{Message: "你没有输入任何内容"}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=system\tmessage=你没有输入任何内容\n"
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
		Items:          []string{"item.tutorial.old_lantern"},
	}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n"
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

func TestTextRenderer_RenderEscapesFieldSeparators(t *testing.T) {
	// Given
	renderer := TextRenderer{}
	event := SystemMessageEvent{Message: "第一行\n第二行\t反斜杠\\"}

	// When
	got := renderer.Render(event)

	// Then
	want := "event=system\tmessage=第一行\\n第二行\\t反斜杠\\\\\n"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
