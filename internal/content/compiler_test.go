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
	if _, ok := compiled.Server.Items["item.tutorial.old_lantern"]; !ok {
		t.Fatal("expected old lantern in server items")
	}
	if got := compiled.Server.ItemLocations["item.tutorial.old_lantern"]; got != "room.tutorial.start" {
		t.Fatalf("expected old lantern in start room, got %q", got)
	}
	if got := compiled.Server.ItemLocations["item.tutorial.practice_sword"]; got != "room.tutorial.yard" {
		t.Fatalf("expected practice sword in yard, got %q", got)
	}
	north, ok := compiled.Server.Items["item.tutorial.north"]
	if !ok {
		t.Fatal("expected north exit item in server items")
	}
	if len(north.Tags) != 1 || north.Tags[0].Exit == nil {
		t.Fatal("expected north item to compile as an exit")
	}
	exit := north.Tags[0].Exit
	if exit.Direction != "north" || exit.TargetRoomID != "room.tutorial.yard" {
		t.Fatalf("north exit = %#v", exit)
	}
}

func TestCompile_rejectsExitNameWithoutTarget(t *testing.T) {
	source := testContentSource()
	source.Items = append(source.Items, ItemSource{
		ID:             "item.tutorial.east",
		InnerNameKey:   "item.tutorial.east.inner_name",
		DisplayNameKey: "item.tutorial.east.name",
		InitialRoom:    "room.tutorial.start",
		Tags:           []SourceTag{{ID: TagExit}},
	})
	source.Text["item.tutorial.east.inner_name"] = "east"

	if _, err := Compile(source); err == nil {
		t.Fatal("expected missing exit target to fail compilation")
	}
}

func TestCompile_rejectsDuplicateStandardDirectionInRoom(t *testing.T) {
	source := testContentSource()
	source.Items = append(source.Items, ItemSource{
		ID:             "item.tutorial.second_north",
		InnerNameKey:   "item.tutorial.second_north.inner_name",
		DisplayNameKey: "item.tutorial.second_north.name",
		InitialRoom:    "room.tutorial.start",
		Tags: []SourceTag{{
			ID:     TagExit,
			Params: map[string]string{"target_room_id": "room.tutorial.yard"},
		}},
	})
	source.Text["item.tutorial.second_north.inner_name"] = "north"

	if _, err := Compile(source); err == nil {
		t.Fatal("expected duplicate north exits to fail compilation")
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
	lanternIndex := 0
	for index, item := range source.Items {
		if item.ID == "item.tutorial.old_lantern" {
			lanternIndex = index
			break
		}
	}
	source.Items[lanternIndex].Aliases = []TextKey{
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

func TestCompile_lockableTag_createsLockableInstance(t *testing.T) {
	source := testContentSource()
	source.Items = append(source.Items, ItemSource{
		ID: "item.test.locked_door", DisplayNameKey: "dn.lock", InnerNameKey: "in.lock",
		DescriptionKey: "dd.lock", InitialRoom: "room.tutorial.start",
		Tags: []SourceTag{
			{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.yard"}},
			{ID: TagLockable, Params: map[string]string{"key_item_id": "item.tutorial.old_lantern"}},
		},
	})
	source.Text["dn.lock"] = "锁着的门"
	source.Text["in.lock"] = "locked door"

	compiled, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	door, ok := compiled.Server.Items["item.test.locked_door"]
	if !ok {
		t.Fatal("expected locked door in compiled output")
	}
	lockableFound := false
	for _, tag := range door.Tags {
		if tag.Lockable != nil && tag.Lockable.KeyItemID == "item.tutorial.old_lantern" {
			lockableFound = true
		}
	}
	if !lockableFound {
		t.Fatal("expected lockable tag with key_item_id=item.tutorial.old_lantern")
	}
}

func TestCompile_lockableTag_rejectsMissingKey(t *testing.T) {
	source := testContentSource()
	source.Items = append(source.Items, ItemSource{
		ID: "item.test.bad_lock", DisplayNameKey: "bl", InnerNameKey: "bl",
		DescriptionKey: "bl", InitialRoom: "room.tutorial.start",
		Tags: []SourceTag{
			{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.yard"}},
			{ID: TagLockable, Params: map[string]string{}},
		},
	})

	if _, err := Compile(source); err == nil {
		t.Fatal("expected lockable without key_item_id to fail")
	}
}

func TestTutorialSource_compilesCurrentTinyWorldFixture(t *testing.T) {
	source := TutorialSource()
	compiled, err := Compile(source)
	if err != nil {
		t.Fatal(err)
	}
	if compiled.Server.StartRoomID != "room.tutorial.hall" {
		t.Fatalf("expected tutorial start room, got %q", compiled.Server.StartRoomID)
	}
	if len(compiled.Server.Rooms) != 5 {
		t.Fatalf("expected 5 rooms, got %d", len(compiled.Server.Rooms))
	}
	if len(compiled.Server.Items) != 12 {
		t.Fatalf("expected 12 items (8 exits + 3 game items + 1 chest), got %d", len(compiled.Server.Items))
	}
	exitTargets := map[ItemID]RoomID{
		"item.hall.north":          "room.tutorial.item_yard",
		"item.hall.east":           "room.tutorial.lock_hall",
		"item.hall.portal":         "room.tutorial.quest_start",
		"item.yard.south":          "room.tutorial.hall",
		"item.lock_hall.west":      "room.tutorial.hall",
		"item.lock_hall.east":      "room.tutorial.lock_chamber",
		"item.lock_chamber.west":   "room.tutorial.lock_hall",
		"item.quest_start.portal":  "room.tutorial.hall",
	}
	for itemID, targetRoomID := range exitTargets {
		item, ok := compiled.Server.Items[itemID]
		if !ok || len(item.Tags) == 0 {
			t.Fatalf("expected %s to compile with tags", itemID)
		}
		exitFound := false
		for _, tag := range item.Tags {
			if tag.Exit != nil {
				exitFound = true
				if tag.Exit.TargetRoomID != targetRoomID {
					t.Fatalf("%s target = %q, want %q", itemID, tag.Exit.TargetRoomID, targetRoomID)
				}
				break
			}
		}
		if !exitFound {
			t.Fatalf("expected %s to compile with an exit tag", itemID)
		}
	}
	if got := compiled.Server.ItemLocations["item.tutorial.old_lantern"]; got != "room.tutorial.lock_hall" {
		t.Fatalf("expected old lantern in lock_hall, got %q", got)
	}
	if got := compiled.Client.Text[compiled.Client.RoomNames["room.tutorial.hall"]]; got != "教学大厅" {
		t.Fatalf("expected hall text, got %q", got)
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
			},
			{
				ID:             "room.tutorial.yard",
				NameKey:        "room.tutorial.yard.name",
				DescriptionKey: "room.tutorial.yard.description",
			},
		},
		Items: []ItemSource{
			{
				ID:             "item.tutorial.north",
				DisplayNameKey: "item.tutorial.north.name",
				InnerNameKey:   "item.tutorial.north.inner_name",
				DescriptionKey: "item.tutorial.north.description",
				Tags: []SourceTag{{
					ID:     TagExit,
					Params: map[string]string{"target_room_id": "room.tutorial.yard"},
				}},
				InitialRoom: "room.tutorial.start",
			},
			{
				ID:             "item.tutorial.south",
				DisplayNameKey: "item.tutorial.south.name",
				InnerNameKey:   "item.tutorial.south.inner_name",
				DescriptionKey: "item.tutorial.south.description",
				Tags: []SourceTag{{
					ID:     TagExit,
					Params: map[string]string{"target_room_id": "room.tutorial.start"},
				}},
				InitialRoom: "room.tutorial.yard",
			},
			{
				ID:             "item.tutorial.old_lantern",
				DisplayNameKey: "item.tutorial.old_lantern.name",
				InnerNameKey:   "item.tutorial.old_lantern.inner_name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				InitialRoom:    "room.tutorial.start",
				Tags:           []SourceTag{{ID: TagCarryable}},
			},
			{
				ID:             "item.tutorial.practice_sword",
				DisplayNameKey: "item.tutorial.practice_sword.name",
				InnerNameKey:   "item.tutorial.practice_sword.inner_name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				InitialRoom:    "room.tutorial.yard",
				Tags:           []SourceTag{{ID: TagCarryable}},
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
			"item.tutorial.north.name":                            "北方",
			"item.tutorial.north.inner_name":                      "north",
			"item.tutorial.north.description":                     "北方的道路。",
			"item.tutorial.south.name":                            "南方",
			"item.tutorial.south.inner_name":                      "south",
			"item.tutorial.south.description":                     "南方的道路。",
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
