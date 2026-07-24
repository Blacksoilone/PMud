package world

import (
	"slices"
	"testing"

	"PMud/internal/content"
)

func TestWorld_NewFromSnapshotPreservesTutorialBehavior(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	game := NewFromSnapshot(compiled.Server, compiled.Client)
	playerID := PlayerID("player.local")

	observation, ok := game.Look(game.StartRoom())
	if !ok {
		t.Fatal("expected start room to exist")
	}
	if observation.Name != "教学大厅" {
		t.Fatalf("expected start room name, got %q", observation.Name)
	}
	// old_lantern is in lock_hall, not in start room
	if slices.Contains(observation.Items, "旧油灯") {
		t.Fatalf("expected no old lantern in start room, got %v", observation.Items)
	}
	// move north to item_yard
	nextRoom, ok := game.Move(game.StartRoom(), "north")
	if !ok {
		t.Fatal("expected north movement to work")
	}
	if nextRoom != "room.tutorial.item_yard" {
		t.Fatalf("expected item_yard, got %q", nextRoom)
	}
	// get old lantern from lock_hall
	itemID, ok, _ := game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)
	if !ok {
		t.Fatal("expected to get old lantern from lock_hall")
	}
	if itemID != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", itemID)
	}
	if !slices.Contains(game.Inventory(playerID), "旧油灯") {
		t.Fatalf("expected old lantern in inventory, got %v", game.Inventory(playerID))
	}
}

func TestWorldMoveUsesExitItems(t *testing.T) {
	game := New()
	startRoom := game.StartRoom()
	nextRoom, moved := game.Move(startRoom, "north")
	if !moved || nextRoom != "room.tutorial.item_yard" {
		t.Fatalf("Move north = %q, %v; want item_yard, true", nextRoom, moved)
	}
}

func TestWorldMoveUsesNamedPortalExit(t *testing.T) {
	game := New()
	startRoom := game.StartRoom()
	nextRoom, moved := game.Move(startRoom, "portal")
	if !moved || nextRoom != "room.tutorial.quest_start" {
		t.Fatalf("Move portal = %q, %v; want quest_start, true", nextRoom, moved)
	}
}

func TestWorldLookSeparatesExitsFromOrdinaryItems(t *testing.T) {
	game := New()
	startRoom := game.StartRoom()
	observation, ok := game.Look(startRoom)
	_, gotExit, _ := game.GetItem(startRoom, "item.hall.north", "player.local")
	if !ok {
		t.Fatal("expected room observation")
	}
	if !slices.Contains(observation.Exits, "north") {
		t.Fatalf("exits = %v, want north", observation.Exits)
	}
	if got := observation.Neighbors["north"]; got != "room.tutorial.item_yard" {
		t.Fatalf("north neighbor = %q, want room.tutorial.item_yard", got)
	}
	if slices.Contains(observation.Items, "北方通路") {
		t.Fatalf("ordinary items include exit: %v", observation.Items)
	}
	if gotExit {
		t.Fatal("expected exit item not to be gettable")
	}
}

func TestWorld_ItemMovesBetweenRoomAndInventory(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	lanternRoom := RoomID("room.tutorial.lock_hall")

	itemID, ok, _ := game.GetItem(lanternRoom, "item.tutorial.old_lantern", playerID)
	if !ok {
		t.Fatal("expected to get old lantern")
	}
	observation, ok := game.Look(lanternRoom)
	if !ok {
		t.Fatal("expected lock_hall to exist")
	}
	if slices.Contains(observation.Items, "旧油灯") {
		t.Fatal("expected old lantern to leave the room after get")
	}
	if !slices.Contains(game.Inventory(playerID), "旧油灯") {
		t.Fatal("expected old lantern to enter player inventory after get")
	}

	droppedItemID, ok := game.DropInventoryItem(lanternRoom, "item.tutorial.old_lantern", playerID)
	if !ok {
		t.Fatal("expected to drop old lantern")
	}
	if droppedItemID != itemID {
		t.Fatalf("expected dropped item %q, got %q", itemID, droppedItemID)
	}
	observation, ok = game.Look(lanternRoom)
	if !ok {
		t.Fatal("expected lock_hall to exist")
	}
	if !slices.Contains(observation.Items, "旧油灯") {
		t.Fatal("expected old lantern to return to the room after drop")
	}
	if slices.Contains(game.Inventory(playerID), "旧油灯") {
		t.Fatal("expected old lantern to leave player inventory after drop")
	}
	if itemID == "" {
		t.Fatal("expected item id to be non-empty")
	}
}

func TestWorld_ExamineItemFindsItemInCurrentRoom(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")

	observation, ok := game.ExamineItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)

	if !ok {
		t.Fatal("expected to examine old lantern in lock_hall")
	}
	if observation.Item != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", observation.Item)
	}
	if observation.Name != "旧油灯" {
		t.Fatalf("expected old lantern name, got %q", observation.Name)
	}
	if observation.Description == "" {
		t.Fatal("expected old lantern description")
	}
}

func TestWorld_ExamineItemFindsItemInInventory(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)

	observation, ok := game.ExamineItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)

	if !ok {
		t.Fatal("expected to examine old lantern in inventory")
	}
	if observation.Item != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", observation.Item)
	}
}

func TestWorld_ExamineItemRejectsInvisibleItem(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")

	_, ok := game.ExamineItem(game.StartRoom(), "item.tutorial.practice_sword", playerID)

	if ok {
		t.Fatal("expected practice sword in another room to be invisible")
	}
}

func TestWorldResolveRoomItemPhrase_matchesDisplayName(t *testing.T) {
	game := New()
	resolution := game.ResolveRoomItemPhrase("room.tutorial.lock_hall", "旧油灯")

	if !resolution.Found {
		t.Fatal("expected old lantern to resolve")
	}
	if resolution.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("item id = %q, want old lantern", resolution.ItemID)
	}
}

func TestWorldResolveRoomItemPhrase_matchesInnerNameSeparatorsCaseInsensitively(t *testing.T) {
	game := New()

	tests := []string{"oldlantern", "old-lantern", "old_lantern", "OLD-LANTERN"}
	for _, phrase := range tests {
		t.Run(phrase, func(t *testing.T) {
			resolution := game.ResolveRoomItemPhrase("room.tutorial.lock_hall", phrase)
			if !resolution.Found {
				t.Fatal("expected old lantern to resolve")
			}
			if resolution.ItemID != "item.tutorial.old_lantern" {
				t.Fatalf("item id = %q, want old lantern", resolution.ItemID)
			}
		})
	}
}

func TestWorldResolveRoomItemPhrase_matchesAliasSeparatorsCaseInsensitively(t *testing.T) {
	game := New()

	tests := []string{"jiuyoudeng", "jiu-youdeng", "jiu_youdeng", "JIU_YOUDENG"}
	for _, phrase := range tests {
		t.Run(phrase, func(t *testing.T) {
			resolution := game.ResolveRoomItemPhrase("room.tutorial.lock_hall", phrase)
			if !resolution.Found {
				t.Fatal("expected old lantern to resolve")
			}
			if resolution.ItemID != "item.tutorial.old_lantern" {
				t.Fatalf("item id = %q, want old lantern", resolution.ItemID)
			}
		})
	}
}

func TestWorldResolveRoomItemPhrase_matchesPracticeSwordPinyinAlias(t *testing.T) {
	game := New()
	resolution := game.ResolveRoomItemPhrase("room.tutorial.item_yard", "lianximujian")

	if !resolution.Found {
		t.Fatal("expected practice sword to resolve")
	}
	if resolution.ItemID != "item.tutorial.practice_sword" {
		t.Fatalf("item id = %q, want practice sword", resolution.ItemID)
	}
}

func TestWorldResolveInventoryItemPhrase_matchesOnlyInventory(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)

	inventoryResolution := game.ResolveInventoryItemPhrase(playerID, "旧油灯")
	roomResolution := game.ResolveRoomItemPhrase("room.tutorial.lock_hall", "旧油灯")

	if !inventoryResolution.Found {
		t.Fatal("expected old lantern to resolve in inventory")
	}
	if inventoryResolution.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("item id = %q, want old lantern", inventoryResolution.ItemID)
	}
	if roomResolution.Found {
		t.Fatalf("expected old lantern not to resolve in room after pickup, got %#v", roomResolution)
	}
}

func TestWorldResolveVisibleItemPhrase_matchesRoomAndInventory(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)

	inventoryResolution := game.ResolveVisibleItemPhrase("room.tutorial.lock_hall", playerID, "旧油灯")
	yardResolution := game.ResolveVisibleItemPhrase("room.tutorial.item_yard", playerID, "练习木剑")

	if !inventoryResolution.Found || inventoryResolution.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("inventory visible resolution = %#v, want old lantern", inventoryResolution)
	}
	if !yardResolution.Found || yardResolution.ItemID != "item.tutorial.practice_sword" {
		t.Fatalf("yard visible resolution = %#v, want practice sword", yardResolution)
	}
}

func TestWorldResolveRoomItemPhrase_reportsAmbiguityOnlyWithinRoom(t *testing.T) {
	game := New()
	game.store.Add(&Entity{
		ID: "item.tutorial.second_lantern", NameKey: "item.tutorial.second_lantern.name", DescriptionKey: "item.tutorial.second_lantern.description",
		Name: "旧油灯", Description: "另一盏旧油灯。",
		Aliases: []string{"jiuyoudeng"},
		Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		Item:    &ItemData{Weight: 0, Volume: 0},
	})
	game.store.PlaceInRoom("item.tutorial.second_lantern", "room.tutorial.lock_hall")
	game.store.Add(&Entity{
		ID: "item.tutorial.distant_lantern", NameKey: "item.tutorial.distant_lantern.name", DescriptionKey: "item.tutorial.distant_lantern.description",
		Name: "旧油灯", Description: "远处的旧油灯。",
		Aliases: []string{"jiuyoudeng"},
		Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		Item:    &ItemData{Weight: 0, Volume: 0},
	})
	game.store.PlaceInRoom("item.tutorial.distant_lantern", "room.tutorial.item_yard")

	lockHallResolution := game.ResolveRoomItemPhrase("room.tutorial.lock_hall", "旧油灯")
	yardResolution := game.ResolveRoomItemPhrase("room.tutorial.item_yard", "旧油灯")

	wantAmbiguous := []ItemID{"item.tutorial.old_lantern", "item.tutorial.second_lantern"}
	if lockHallResolution.Found {
		t.Fatalf("expected ambiguous lock_hall resolution, got found %#v", lockHallResolution)
	}
	if !slices.Equal(lockHallResolution.AmbiguousItemIDs, wantAmbiguous) {
		t.Fatalf("ambiguous ids = %#v, want %#v", lockHallResolution.AmbiguousItemIDs, wantAmbiguous)
	}
	if !yardResolution.Found || yardResolution.ItemID != "item.tutorial.distant_lantern" {
		t.Fatalf("yard resolution = %#v, want distant lantern only", yardResolution)
	}
}

func TestWorldGetItemAllowsCarryableExit(t *testing.T) {
	game := New()
	startRoom := game.StartRoom()
	game.store.Add(&Entity{
		ID: "item.hall.test_portal", Name: "传送门",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "test_portal", "target": "room.tutorial.item_yard"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
		Item: &ItemData{Weight: 0, Volume: 0},
	})
	game.store.PlaceInRoom("item.hall.test_portal", startRoom)

	itemID, got, _ := game.GetItem(startRoom, "item.hall.test_portal", "player.local")
	if !got || itemID != "item.hall.test_portal" {
		t.Fatalf("GetItem carryable exit = %q, %v", itemID, got)
	}
}

func TestWorldDropInventoryItemRejectsDuplicateExitDirection(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	startRoom := game.StartRoom()
	game.store.Add(&Entity{
		ID: "item.tutorial.moving_north", Name: "另一个北方",
		Exit: &ExitData{Direction: "north", TargetRoomID: "room.tutorial.lock_hall"},
		Item: &ItemData{Weight: 0, Volume: 0},
	})
	game.containerContents[PlayerContainerID(playerID)] = append(game.containerContents[PlayerContainerID(playerID)], "item.tutorial.moving_north")

	itemID, dropped := game.DropInventoryItem(startRoom, "item.tutorial.moving_north", playerID)
	if dropped || itemID != "" {
		t.Fatalf("DropInventoryItem duplicate north = %q, %v; want empty, false", itemID, dropped)
	}
	if !slices.Contains(game.Inventory(playerID), "另一个北方") {
		t.Fatal("rejected exit drop removed item from inventory")
	}
}

func TestTwoPlayersHaveIndependentInventories(t *testing.T) {
	game := New()

	startA, ok := game.EnterWorld("player.a")
	if !ok {
		t.Fatal("expected player A to enter world")
	}
	startB, ok := game.EnterWorld("player.b")
	if !ok {
		t.Fatal("expected player B to enter world")
	}

	if startA != game.StartRoom() || startB != game.StartRoom() {
		t.Fatal("both players should start in start room")
	}

	// A picks up lantern from lock_hall (physics-based; no movement needed for this world-level test)
	_, ok, _ = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.a")
	if !ok {
		t.Fatal("A should get the lantern")
	}

	// B should not see lantern in lock_hall (checked by same item state)
	// after A picked it up, it's in A's inventory — no longer visible from any room
	_, ok, _ = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.b")
	if ok {
		t.Fatal("B should not be able to get lantern that A has")
	}

	// A has lantern in inventory
	if !slices.Contains(game.Inventory("player.a"), "旧油灯") {
		t.Fatal("A should have the lantern in inventory")
	}

	// B does not have lantern
	if slices.Contains(game.Inventory("player.b"), "旧油灯") {
		t.Fatal("B should not have the lantern")
	}
}

func TestMoveBlocked_blocksPlayerStaysInPlace(t *testing.T) {
	game := New()
	game.EnterWorld("player.a")

	_, ok := game.MovePlayer("player.a", "nowhere")
	if ok {
		t.Fatal("expected nowhere direction to fail")
	}
	room, _ := game.PlayerRoom("player.a")
	if room != game.StartRoom() {
		t.Fatalf("expected to stay in start room, got %q", room)
	}
}

func TestTwoPlayersMoveIndependently(t *testing.T) {
	game := New()

	game.EnterWorld("player.a")
	game.EnterWorld("player.b")

	// A: start in hall → move east to lock_hall → get lantern → move east to lock_chamber (unlocked with lantern)
	_, ok := game.MovePlayer("player.a", "east")
	if !ok {
		t.Fatal("A should be able to move east to lock_hall")
	}
	_, ok, _ = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.a")
	if !ok {
		t.Fatal("A should be able to get old lantern in lock_hall")
	}
	roomA, ok := game.MovePlayer("player.a", "east")
	if !ok {
		t.Fatal("A should be able to move east with lantern")
	}
	if roomA != "room.tutorial.lock_chamber" {
		t.Fatalf("A should be in lock_chamber, got %q", roomA)
	}

	// B should still be in start (hall)
	roomB, ok := game.PlayerRoom("player.b")
	if !ok {
		t.Fatal("player B should exist")
	}
	if roomB != game.StartRoom() {
		t.Fatalf("B should still be in start room, got %q", roomB)
	}

	// Only A in lock_chamber
	chamberPlayers := game.PlayersInRoom("room.tutorial.lock_chamber")
	if !slices.Contains(chamberPlayers, PlayerID("player.a")) {
		t.Fatal("A should be in lock_chamber")
	}
	if slices.Contains(chamberPlayers, PlayerID("player.b")) {
		t.Fatal("B should not be in lock_chamber")
	}

	// Only B in hall
	startPlayers := game.PlayersInRoom(game.StartRoom())
	if !slices.Contains(startPlayers, PlayerID("player.b")) {
		t.Fatal("B should be in hall")
	}
	if slices.Contains(startPlayers, PlayerID("player.a")) {
		t.Fatal("A should not be in hall")
	}
}

func TestTwoPlayersTrackedByLeaveWorld(t *testing.T) {
	game := New()

	game.EnterWorld("player.a")
	game.EnterWorld("player.b")

	if game.PlayerCount() != 2 {
		t.Fatalf("player count = %d, want 2", game.PlayerCount())
	}

	game.LeaveWorld("player.a")

	if game.PlayerCount() != 1 {
		t.Fatalf("player count = %d, want 1", game.PlayerCount())
	}

	// B is unaffected
	if _, ok := game.PlayerRoom("player.b"); !ok {
		t.Fatal("B should still exist after A left")
	}
}

func TestTagRegistry_defineAndLookup(t *testing.T) {
	w := New()
	def, ok := w.TagDefinition("tag.exit")
	if !ok {
		t.Fatal("expected tag.exit to be registered")
	}
	if def.Description == "" {
		t.Fatal("expected non-empty description")
	}
	// 验证字段 schema
	var hasTarget bool
	for _, f := range def.Fields {
		if f.Name == "target" {
			hasTarget = true
			if f.Type != TagFieldRef {
				t.Fatalf("target field type = %q, want ref", f.Type)
			}
			if !f.Required {
				t.Fatal("target field should be required")
			}
		}
	}
	if !hasTarget {
		t.Fatal("exit tag missing target field")
	}
}

func TestTagRegistry_rejectsDuplicateID(t *testing.T) {
	w := New()
	err := w.RegisterTag(TagDefinition{
		ID:          "tag.exit", // already registered
		Description: "dup",
		Scopes:      []TagScope{TagScopeItem},
	})
	if err == nil {
		t.Fatal("expected error for duplicate tag ID")
	}
}

func TestTagRegistry_listDefinitions(t *testing.T) {
	w := New()
	defs := w.TagDefinitions()
	if _, ok := defs["tag.exit"]; !ok {
		t.Fatal("expected tag.exit in definitions")
	}
	if _, ok := defs["tag.carryable"]; !ok {
		t.Fatal("expected tag.carryable in definitions")
	}
}

func TestEntityTag_dataAccessibleViaStore(t *testing.T) {
	w := New()
	w.store.Add(&Entity{
		ID: "item.test_exit",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.test_target"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
		Item: &ItemData{Weight: 0, Volume: 0},
		Exit: &ExitData{Direction: "north", TargetRoomID: "room.test_target"},
	})
	w.store.Add(&Entity{
		ID: "item.test_carryable",
		Tags: []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		Item: &ItemData{Weight: 1, Volume: 1},
	})
	w.store.Add(&Entity{
		ID: "item.not_carryable", Name: "not carryable",
		Item: &ItemData{Weight: 1, Volume: 1},
	})

	if !w.store.Tag("item.test_exit", "tag.exit") {
		t.Fatal("expected test_exit to have tag.exit")
	}
	if !w.store.Tag("item.test_exit", "tag.carryable") {
		t.Fatal("expected test_exit to have tag.carryable")
	}
	ed := w.store.Exit("item.test_exit")
	if ed == nil || ed.Direction != "north" || ed.TargetRoomID != "room.test_target" {
		t.Fatalf("exit data = %+v, want north→test_target", ed)
	}
	if !w.itemIsCarryable("item.test_carryable") {
		t.Fatal("expected carryable item")
	}
	if w.itemIsCarryable("item.not_carryable") {
		t.Fatal("expected non-carryable item")
	}
}

func TestLightableTag_definedAndAccessible(t *testing.T) {
	w := New()
	if !w.store.Tag("item.tutorial.old_lantern", "tag.lightable") {
		t.Fatal("expected old lantern to have tag.lightable")
	}
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	w2 := NewFromSnapshot(compiled.Server, compiled.Client)
	if !w2.store.Tag("item.tutorial.old_lantern", "tag.lightable") {
		t.Fatal("expected old lantern to have tag.lightable in compiled world")
	}
}

func TestContainerTag_definedAndAccessible(t *testing.T) {
	w := New()
	w.store.Add(&Entity{
		ID: "item.test_backpack", Name: "背包",
		Tags: []TagInstance{
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		},
		Item: &ItemData{Weight: 1, Volume: 1},
	})
	if !w.itemIsContainer("item.test_backpack") {
		t.Fatal("expected backpack to be container")
	}
	capacity := w.containerCapacity("item.test_backpack")
	if capacity != 5 {
		t.Fatalf("capacity = %d, want 5", capacity)
	}
}

func TestLightableTag_compilePipeline(t *testing.T) {
	// 验证 compileTags 能正常处理 lightable SourceTag
	source := content.ContentSource{
		StartRoomID: "room.test",
		Rooms:       []content.RoomSource{{ID: "room.test", NameKey: "rk", DescriptionKey: "dk"}},
		Items: []content.ItemSource{{
			ID: "item.test.candle", DisplayNameKey: "dn", InnerNameKey: "in",
			DescriptionKey: "dd", InitialRoom: "room.test",
			Tags: []content.SourceTag{{ID: content.TagLightable}},
		}},
		Text: map[content.TextKey]string{"dn": "蜡烛", "in": "candle", "dd": "一支蜡烛。"},
	}
	compiled, err := content.Compile(source)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	serverItem, ok := compiled.Server.Items["item.test.candle"]
	if !ok {
		t.Fatal("expected candle in compiled output")
	}
	if !serverItem.Tags[0].Lightable {
		t.Fatal("expected compiled candle to be lightable")
	}
}

func TestContainerTag_compilePipeline(t *testing.T) {
	source := content.ContentSource{
		StartRoomID: "room.test",
		Rooms:       []content.RoomSource{{ID: "room.test", NameKey: "rk", DescriptionKey: "dk"}},
		Items: []content.ItemSource{{
			ID: "item.test.box", DisplayNameKey: "dn", InnerNameKey: "in",
			DescriptionKey: "dd", InitialRoom: "room.test",
			Tags: []content.SourceTag{
				{ID: content.TagContainer, Params: map[string]string{"capacity": "3"}},
			},
		}},
		Text: map[content.TextKey]string{"dn": "箱子", "in": "box", "dd": "一个箱子。"},
	}
	compiled, err := content.Compile(source)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	serverItem, ok := compiled.Server.Items["item.test.box"]
	if !ok {
		t.Fatal("expected box in compiled output")
	}
	if serverItem.Tags[0].Container == nil || serverItem.Tags[0].Container.Capacity != 3 {
		t.Fatalf("Container.Capacity = %d, want 3", serverItem.Tags[0].Container.Capacity)
	}

	// 再验证从 ServerSnapshot → world 也能正确转换
	w := NewFromSnapshot(compiled.Server, compiled.Client)
	ent := w.store.Get("item.test.box")
	if ent == nil {
		t.Fatal("expected box in store")
	}
	var params map[string]any
	var found bool
	for _, inst := range ent.Tags {
		if inst.DefinitionID == "tag.container" {
			params = inst.Params
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected box to have tag.container in world")
	}
	capacity, ok := params["capacity"].(int)
	if !ok || capacity != 3 {
		t.Fatalf("capacity = %v, want 3", params["capacity"])
	}
}

func TestContainer_OpenClose(t *testing.T) {
	w := New()
	w.store.Add(&Entity{
		ID: "item.box", Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 2}}},
		Item: &ItemData{},
	})
	w.store.PlaceInRoom("item.box", w.startRoom)
	if w.ContainerIsOpen("item.box") {
		t.Fatal("container should start closed")
	}
	if !w.OpenContainer("item.box") {
		t.Fatal("expected open to succeed")
	}
	if !w.ContainerIsOpen("item.box") {
		t.Fatal("container should be open after open")
	}
	if !w.CloseContainer("item.box") {
		t.Fatal("expected close to succeed")
	}
	if w.ContainerIsOpen("item.box") {
		t.Fatal("container should be closed after close")
	}
	// 非容器物品拒绝操作
	w.store.Add(&Entity{ID: "item.rock", Name: "石头", Item: &ItemData{}})
	if w.OpenContainer("item.rock") {
		t.Fatal("non-container should not open")
	}
}

func TestContainer_PutAndGet(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.store.Add(&Entity{ID: playerID, Name: string(playerID), Player: &PlayerData{}})
	w.store.PlaceInRoom(playerID, w.startRoom)
	// 准备容器
	w.store.Add(&Entity{
		ID: "item.box", Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 2}}},
		Item: &ItemData{},
	})
	w.store.PlaceInRoom("item.box", w.startRoom)
	w.OpenContainer("item.box")
	// 准备物品（在玩家背包中）
	w.store.Add(&Entity{ID: "item.apple", Name: "苹果", Item: &ItemData{}})
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], "item.apple")
	// put apple in box
	err := w.PutItemInContainer("item.apple", "item.box", playerID)
	if err != nil {
		t.Fatalf("PutItemInContainer: %v", err)
	}
	// 验证苹果在容器中
	contents := w.ContainerContents("item.box")
	if len(contents) != 1 || contents[0] != "item.apple" {
		t.Fatalf("contents = %v, want [item.apple]", contents)
	}
	// get apple from box
	err = w.GetItemFromContainer("item.box", "item.apple", playerID)
	if err != nil {
		t.Fatalf("GetItemFromContainer: %v", err)
	}
	// 验证苹果回到了玩家背包
	inventory := w.containerContents[PlayerContainerID(playerID)]
	if len(inventory) != 1 || inventory[0] != "item.apple" {
		t.Fatalf("inventory = %v, want [item.apple]", inventory)
	}
}

func TestContainer_CapacityLimit(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.store.Add(&Entity{ID: playerID, Name: string(playerID), Player: &PlayerData{}})
	w.store.PlaceInRoom(playerID, w.startRoom)
	w.store.Add(&Entity{
		ID: "item.box", Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 1}}},
		Item: &ItemData{},
	})
	w.store.PlaceInRoom("item.box", w.startRoom)
	w.OpenContainer("item.box")
	w.store.Add(&Entity{ID: "item.a", Name: "A", Item: &ItemData{}})
	w.store.Add(&Entity{ID: "item.b", Name: "B", Item: &ItemData{}})
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], "item.a", "item.b")
	if err := w.PutItemInContainer("item.a", "item.box", playerID); err != nil {
		t.Fatalf("first put: %v", err)
	}
	err := w.PutItemInContainer("item.b", "item.box", playerID)
	if err == nil {
		t.Fatal("expected capacity error on second put")
	}
}

func TestContainer_NestingRule(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.store.Add(&Entity{ID: playerID, Name: string(playerID), Player: &PlayerData{}})
	w.store.PlaceInRoom(playerID, w.startRoom)
	// 便携容器（收纳袋）— carryable + container
	w.store.Add(&Entity{
		ID: "item.bag", Name: "收纳袋",
		Tags: []TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		},
		Item: &ItemData{},
	})
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], "item.bag")
	// 另一个 carryable 容器 — 应该不能放入收纳袋
	w.store.Add(&Entity{
		ID: "item.pouch", Name: "小袋",
		Tags: []TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 3}},
		},
		Item: &ItemData{},
	})
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], "item.pouch")
	// 打开收纳袋
	w.OpenContainer("item.bag")
	// 尝试把小袋放进收纳袋
	err := w.PutItemInContainer("item.pouch", "item.bag", playerID)
	if err == nil {
		t.Fatal("expected nesting rejection: carryable container in carryable container")
	}
}

func TestContainer_ClosedRejectsPut(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.store.Add(&Entity{ID: playerID, Name: string(playerID), Player: &PlayerData{}})
	w.store.PlaceInRoom(playerID, w.startRoom)
	w.store.Add(&Entity{
		ID: "item.box", Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}}},
		Item: &ItemData{},
	})
	w.store.PlaceInRoom("item.box", w.startRoom)
	// 不打开
	w.store.Add(&Entity{ID: "item.apple", Name: "苹果", Item: &ItemData{}})
	w.containerContents[PlayerContainerID(playerID)] = append(w.containerContents[PlayerContainerID(playerID)], "item.apple")
	err := w.PutItemInContainer("item.apple", "item.box", playerID)
	if err == nil {
		t.Fatal("expected rejection: container is closed")
	}
	// 从关闭的容器中取也不行
	err = w.GetItemFromContainer("item.box", "item.apple", playerID)
	if err == nil {
		t.Fatal("expected rejection: container is closed")
	}
}

func TestCarryableContainer_WeightAndVolume(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.EnterWorld(playerID)

	// 准备便携容器
	w.store.Add(&Entity{
		ID: "item.bag", Name: "收纳袋",
		Tags: []TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		},
		Item: &ItemData{Weight: 2, Volume: 1},
	})
	w.store.PlaceInRoom("item.bag", w.PlayerCurrentRoom(playerID))

	// 准备内容物
	w.store.Add(&Entity{
		ID: "item.coin", Name: "金币",
		Tags: []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		Item: &ItemData{Weight: 1, Volume: 1},
	})
	w.store.PlaceInRoom("item.coin", w.PlayerCurrentRoom(playerID))

	// 1. 拿起收纳袋 — 体积+1，重量+2
	_, ok, _ := w.GetItem(w.PlayerCurrentRoom(playerID), "item.bag", playerID)
	if !ok {
		t.Fatal("expected to pick up carryable container")
	}
	if wt, _ := w.PlayerWeightRatio(playerID); wt != 2 {
		t.Fatalf("weight after pickup = %d, want 2", wt)
	}
	if _, vol := w.PlayerVolumeRatio(playerID); vol != 10 {
		t.Fatalf("max volume = %d, want 10 (player default)", vol)
	}
	if cur, _ := w.PlayerVolumeRatio(playerID); cur != 1 {
		t.Fatalf("volume after pickup = %d, want 1 (pouch volume, not contents)", cur)
	}

	// 2. 捡起金币，打开袋子，把金币放进去
	_, ok, _ = w.GetItem(w.PlayerCurrentRoom(playerID), "item.coin", playerID)
	if !ok {
		t.Fatal("expected to pick up coin")
	}
	w.OpenContainer("item.bag")
	err := w.PutItemInContainer("item.coin", "item.bag", playerID)
	if err != nil {
		t.Fatalf("PutItemInContainer: %v", err)
	}

	// 3. 重量增加了（含内容物），体积不变（内容物不占背包体积）
	if wt, _ := w.PlayerWeightRatio(playerID); wt != 3 {
		t.Fatalf("weight after putting coin in bag = %d, want 3 (bag 2 + coin 1)", wt)
	}
	if cur, _ := w.PlayerVolumeRatio(playerID); cur != 1 {
		t.Fatalf("volume with coin in bag = %d, want 1 (only bag volume)", cur)
	}

	// 4. 从袋子里取出金币 — 重量不变（仍在背包里），体积增加（金币本身占体积）
	_ = w.GetItemFromContainer("item.bag", "item.coin", playerID)
	if wt, _ := w.PlayerWeightRatio(playerID); wt != 3 {
		t.Fatalf("weight after removing coin = %d, want 3 (bag %d + coin %d)", wt, 2, 1)
	}
	// 现在金币在背包里，体积要计入
	if cur, _ := w.PlayerVolumeRatio(playerID); cur != 2 {
		t.Fatalf("volume after removing coin = %d, want 2 (bag %d + coin %d)", cur, 1, 1)
	}

	// 5. 把金币放回袋子 — 重量不变，体积减少
	err = w.PutItemInContainer("item.coin", "item.bag", playerID)
	if err != nil {
		t.Fatalf("PutItemInContainer: %v", err)
	}
	if wt, _ := w.PlayerWeightRatio(playerID); wt != 3 {
		t.Fatalf("weight after putting coin back = %d, want 3", wt)
	}
	if cur, _ := w.PlayerVolumeRatio(playerID); cur != 1 {
		t.Fatalf("volume after putting coin back = %d, want 1 (bag only)", cur)
	}
}

func TestContainer_GetRejectsVolumeOverflow(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.EnterWorld(playerID)
	w.players[playerID] = PlayerEntity{RoomID: w.PlayerCurrentRoom(playerID), MaxVolume: 1, MaxWeight: 1}
	w.items["item.box"] = Item{
		Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 2}}},
	}
	w.itemLocations["item.box"] = RoomItemLocation{RoomID: w.PlayerCurrentRoom(playerID)}
	w.OpenContainer("item.box")
	w.items["item.stone"] = Item{Name: "石头", Volume: 2, Weight: 10}
	w.itemLocations["item.stone"] = ContainerItemLocation{ContainerID: ItemContainerID("item.box")}

	if err := w.GetItemFromContainer("item.box", "item.stone", playerID); err == nil {
		t.Fatal("expected volume overflow to reject container pickup")
	}
	if got := w.itemsInContainer(ItemContainerID("item.box")); len(got) != 1 || got[0] != "item.stone" {
		t.Fatalf("container contents = %v, want stone retained", got)
	}
}
