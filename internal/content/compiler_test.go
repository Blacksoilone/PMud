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
	if got := compiled.Client.ItemDisplayNames["item.tutorial.old_lantern"]; got != "item.tutorial.old_lantern.name" {
		t.Fatalf("expected item display name key, got %q", got)
	}
	if got := compiled.Client.ItemInnerNames["item.tutorial.old_lantern"]; got != "item.tutorial.old_lantern.inner_name" {
		t.Fatalf("expected item inner name key, got %q", got)
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
	if got := compiled.Client.Text["item.tutorial.old_lantern.inner_name"]; got != "old lantern" {
		t.Fatalf("expected item inner name text, got %q", got)
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

func TestCompile_projectsTutorialQuest(t *testing.T) {
	// Given
	source := testContentSource()

	// When
	compiled, err := Compile(source)
	// Then
	if err != nil {
		t.Fatal(err)
	}
	quest, ok := compiled.Server.Quests["quest.tutorial.first_steps"]
	if !ok {
		t.Fatal("missing tutorial quest")
	}
	if quest.NameKey != "quest.tutorial.first_steps.name" {
		t.Fatalf("quest name key = %q", quest.NameKey)
	}
	if len(quest.StageIDs) != 3 {
		t.Fatalf("stage count = %d, want 3", len(quest.StageIDs))
	}
	if quest.StageIDs[0] != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage 0 = %q", quest.StageIDs[0])
	}
	stage := compiled.Server.QuestStages[quest.StageIDs[0]]
	if stage.TextKey != "quest.tutorial.first_steps.stage.get_lantern.text" {
		t.Fatalf("stage text key = %q", stage.TextKey)
	}
	if len(stage.FinishConditions) != 1 {
		t.Fatalf("condition count = %d, want 1", len(stage.FinishConditions))
	}
	condition := stage.FinishConditions[0]
	if condition.Kind != "got_item" || condition.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("condition = %#v, want got old lantern", condition)
	}
	if got := compiled.Client.Text[quest.NameKey]; got != "教程任务" {
		t.Fatalf("quest name text = %q", got)
	}
	if got := compiled.Client.Text[stage.TextKey]; got != "拿起旧油灯。" {
		t.Fatalf("stage text = %q", got)
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
	if got := compiled.Client.Text[compiled.Client.ItemDisplayNames["item.tutorial.practice_sword"]]; got != "练习木剑" {
		t.Fatalf("expected practice sword text, got %q", got)
	}
	if got := compiled.Client.Text[compiled.Client.ItemInnerNames["item.tutorial.practice_sword"]]; got != "practice sword" {
		t.Fatalf("expected practice sword inner name text, got %q", got)
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
				DisplayNameKey: "item.tutorial.old_lantern.name",
				InnerNameKey:   "item.tutorial.old_lantern.inner_name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				InitialRoom:    "room.tutorial.start",
			},
			{
				ID:             "item.tutorial.practice_sword",
				DisplayNameKey: "item.tutorial.practice_sword.name",
				InnerNameKey:   "item.tutorial.practice_sword.inner_name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				InitialRoom:    "room.tutorial.yard",
			},
		},
		Quests: []QuestSource{
			{
				ID:      "quest.tutorial.first_steps",
				NameKey: "quest.tutorial.first_steps.name",
				StageIDs: []QuestStageID{
					"quest.tutorial.first_steps.stage.get_lantern",
					"quest.tutorial.first_steps.stage.enter_yard",
					"quest.tutorial.first_steps.stage.examine_sword",
				},
			},
		},
		QuestStages: []QuestStageSource{
			{
				ID:      "quest.tutorial.first_steps.stage.get_lantern",
				TextKey: "quest.tutorial.first_steps.stage.get_lantern.text",
				FinishConditions: []QuestConditionSource{
					{Kind: "got_item", ItemID: "item.tutorial.old_lantern"},
				},
				NextStageID: "quest.tutorial.first_steps.stage.enter_yard",
			},
			{
				ID:      "quest.tutorial.first_steps.stage.enter_yard",
				TextKey: "quest.tutorial.first_steps.stage.enter_yard.text",
				FinishConditions: []QuestConditionSource{
					{Kind: "moved_room", RoomID: "room.tutorial.yard"},
				},
				NextStageID: "quest.tutorial.first_steps.stage.examine_sword",
			},
			{
				ID:      "quest.tutorial.first_steps.stage.examine_sword",
				TextKey: "quest.tutorial.first_steps.stage.examine_sword.text",
				FinishConditions: []QuestConditionSource{
					{Kind: "examined_item", ItemID: "item.tutorial.practice_sword"},
				},
			},
		},
		Text: map[TextKey]string{
			"room.tutorial.start.name":                            "练习场入口",
			"room.tutorial.start.description":                     "这里是练习场的入口。北边传来木剑碰撞的声音。",
			"room.tutorial.yard.name":                             "练习场",
			"room.tutorial.yard.description":                      "几根木桩立在泥地上，地面满是被踩出的脚印。",
			"item.tutorial.old_lantern.name":                      "旧油灯",
			"item.tutorial.old_lantern.inner_name":                "old lantern",
			"item.tutorial.old_lantern.description":               "灯罩上蒙着一层灰，里面还剩一点灯油。",
			"item.tutorial.practice_sword.name":                   "练习木剑",
			"item.tutorial.practice_sword.inner_name":             "practice sword",
			"item.tutorial.practice_sword.description":            "一把被许多人握过的木剑，剑柄已经磨得发亮。",
			"quest.tutorial.first_steps.name":                     "教程任务",
			"quest.tutorial.first_steps.stage.get_lantern.text":   "拿起旧油灯。",
			"quest.tutorial.first_steps.stage.enter_yard.text":    "前往练习场。",
			"quest.tutorial.first_steps.stage.examine_sword.text": "查看练习木剑。",
		},
	}
}
