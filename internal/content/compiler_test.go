package content

import "testing"

func TestCompile_projectsServerSnapshot(t *testing.T) {
	// Given
	source := testContentSource()

	// When
	compiled, err := Compile(source)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if compiled.Server.StartRoomID != "room.tutorial.start" {
		t.Fatalf("expected start room %q, got %q", "room.tutorial.start", compiled.Server.StartRoomID)
	}
	if got := compiled.Server.Rooms["room.tutorial.start"].Exits["north"]; got != "room.tutorial.yard" {
		t.Fatalf("expected north exit to yard, got %q", got)
	}
	if got := compiled.Server.Rooms["room.tutorial.yard"].Exits["south"]; got != "room.tutorial.start" {
		t.Fatalf("expected south exit to start, got %q", got)
	}
	if _, ok := compiled.Server.Items["item.tutorial.old_lantern"]; !ok {
		t.Fatal("expected old lantern in server items")
	}
	if got := compiled.Server.ItemLocations["item.tutorial.old_lantern"]; got != "room.tutorial.start" {
		t.Fatalf("expected old lantern in start room, got %q", got)
	}
	if got := compiled.Server.ItemLocations["item.tutorial.practice_sword"]; got != "room.tutorial.yard" {
		t.Fatalf("expected practice sword in yard, got %q", got)
	}
}

func TestCompile_projectsClientCatalog(t *testing.T) {
	// Given
	source := testContentSource()

	// When
	compiled, err := Compile(source)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if got := compiled.Client.RoomNames["room.tutorial.start"]; got != "room.tutorial.start.name" {
		t.Fatalf("expected room name key, got %q", got)
	}
	if got := compiled.Client.RoomDescriptions["room.tutorial.start"]; got != "room.tutorial.start.description" {
		t.Fatalf("expected room description key, got %q", got)
	}
	if got := compiled.Client.ItemNames["item.tutorial.old_lantern"]; got != "item.tutorial.old_lantern.name" {
		t.Fatalf("expected item name key, got %q", got)
	}
	if got := compiled.Client.ItemDescriptions["item.tutorial.old_lantern"]; got != "item.tutorial.old_lantern.description" {
		t.Fatalf("expected item description key, got %q", got)
	}
	if got := compiled.Client.Text["room.tutorial.start.name"]; got != "练习场入口" {
		t.Fatalf("expected room name text, got %q", got)
	}
	if got := compiled.Client.Text["item.tutorial.old_lantern.name"]; got != "旧油灯" {
		t.Fatalf("expected item name text, got %q", got)
	}
}

func TestCompile_projectsClientItemAliases(t *testing.T) {
	// Given
	source := testContentSource()
	source.Items[0].Aliases = []TextKey{
		"item.tutorial.old_lantern.alias.jiuyoudeng",
		"item.tutorial.old_lantern.alias.old_lantern",
	}
	source.Text["item.tutorial.old_lantern.alias.jiuyoudeng"] = "jiuyoudeng"
	source.Text["item.tutorial.old_lantern.alias.old_lantern"] = "old_lantern"

	// When
	compiled, err := Compile(source)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	got := compiled.Client.ItemAliases["item.tutorial.old_lantern"]
	if len(got) != 2 {
		t.Fatalf("expected 2 item aliases, got %d", len(got))
	}
	if got[0] != "item.tutorial.old_lantern.alias.jiuyoudeng" {
		t.Fatalf("expected first alias key, got %q", got[0])
	}
	if got[1] != "item.tutorial.old_lantern.alias.old_lantern" {
		t.Fatalf("expected second alias key, got %q", got[1])
	}
	if got := compiled.Client.Text[got[0]]; got != "jiuyoudeng" {
		t.Fatalf("expected alias text to be copied, got %q", got)
	}
}

func TestTutorialSource_compilesCurrentTinyWorldFixture(t *testing.T) {
	// Given
	source := TutorialSource()

	// When
	compiled, err := Compile(source)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if compiled.Server.StartRoomID != "room.tutorial.start" {
		t.Fatalf("expected tutorial start room, got %q", compiled.Server.StartRoomID)
	}
	if len(compiled.Server.Rooms) != 2 {
		t.Fatalf("expected 2 rooms, got %d", len(compiled.Server.Rooms))
	}
	if len(compiled.Server.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(compiled.Server.Items))
	}
	if got := compiled.Server.ItemLocations["item.tutorial.old_lantern"]; got != "room.tutorial.start" {
		t.Fatalf("expected old lantern in start room, got %q", got)
	}
	if got := compiled.Client.Text[compiled.Client.RoomNames["room.tutorial.start"]]; got != "练习场入口" {
		t.Fatalf("expected start room text, got %q", got)
	}
	if got := compiled.Client.Text[compiled.Client.ItemNames["item.tutorial.practice_sword"]]; got != "练习木剑" {
		t.Fatalf("expected practice sword text, got %q", got)
	}
}

func testContentSource() ContentSource {
	return ContentSource{
		StartRoomID: "room.tutorial.start",
		Rooms: []RoomSource{
			{
				ID:             "room.tutorial.start",
				NameKey:        "room.tutorial.start.name",
				DescriptionKey: "room.tutorial.start.description",
				Exits: map[Direction]RoomID{
					"north": "room.tutorial.yard",
				},
			},
			{
				ID:             "room.tutorial.yard",
				NameKey:        "room.tutorial.yard.name",
				DescriptionKey: "room.tutorial.yard.description",
				Exits: map[Direction]RoomID{
					"south": "room.tutorial.start",
				},
			},
		},
		Items: []ItemSource{
			{
				ID:             "item.tutorial.old_lantern",
				NameKey:        "item.tutorial.old_lantern.name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				InitialRoom:    "room.tutorial.start",
			},
			{
				ID:             "item.tutorial.practice_sword",
				NameKey:        "item.tutorial.practice_sword.name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				InitialRoom:    "room.tutorial.yard",
			},
		},
		Text: map[TextKey]string{
			"room.tutorial.start.name":                 "练习场入口",
			"room.tutorial.start.description":          "这里是练习场的入口。北边传来木剑碰撞的声音。",
			"room.tutorial.yard.name":                  "练习场",
			"room.tutorial.yard.description":           "几根木桩立在泥地上，地面满是被踩出的脚印。",
			"item.tutorial.old_lantern.name":           "旧油灯",
			"item.tutorial.old_lantern.description":    "灯罩上蒙着一层灰，里面还剩一点灯油。",
			"item.tutorial.practice_sword.name":        "练习木剑",
			"item.tutorial.practice_sword.description": "一把被许多人握过的木剑，剑柄已经磨得发亮。",
		},
	}
}
