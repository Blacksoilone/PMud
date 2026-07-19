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
	itemID, ok := game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", playerID)
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
	_, gotExit := game.GetItem(startRoom, "item.hall.north", "player.local")
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

	itemID, ok := game.GetItem(lanternRoom, "item.tutorial.old_lantern", playerID)
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
	game.items["item.tutorial.second_lantern"] = Item{
		NameKey: "item.tutorial.second_lantern.name", DescriptionKey: "item.tutorial.second_lantern.description",
		Name: "旧油灯", InnerName: "old lantern", Description: "另一盏旧油灯。",
		Aliases: []string{"jiuyoudeng"},
		Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
	}
	game.itemLocations["item.tutorial.second_lantern"] = RoomItemLocation{RoomID: "room.tutorial.lock_hall"}
	game.items["item.tutorial.distant_lantern"] = Item{
		NameKey: "item.tutorial.distant_lantern.name", DescriptionKey: "item.tutorial.distant_lantern.description",
		Name: "旧油灯", InnerName: "old lantern", Description: "远处的旧油灯。",
		Aliases: []string{"jiuyoudeng"},
		Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
	}
	game.itemLocations["item.tutorial.distant_lantern"] = RoomItemLocation{RoomID: "room.tutorial.item_yard"}

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
	game.items["item.hall.test_portal"] = Item{
		Name: "传送门", InnerName: "test_portal",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "test_portal", "target": "room.tutorial.item_yard"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
	}
	game.itemLocations["item.hall.test_portal"] = RoomItemLocation{RoomID: startRoom}

	itemID, got := game.GetItem(startRoom, "item.hall.test_portal", "player.local")
	if !got || itemID != "item.hall.test_portal" {
		t.Fatalf("GetItem carryable exit = %q, %v", itemID, got)
	}
}

func TestWorldDropInventoryItemRejectsDuplicateExitDirection(t *testing.T) {
	game := New()
	playerID := PlayerID("player.local")
	startRoom := game.StartRoom()
	game.items["item.tutorial.moving_north"] = Item{
		Name: "另一个北方", InnerName: "north",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.lock_hall"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
	}
	game.itemLocations["item.tutorial.moving_north"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}

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
	_, ok = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.a")
	if !ok {
		t.Fatal("A should get the lantern")
	}

	// B should not see lantern in lock_hall (checked by same item state)
	// after A picked it up, it's in A's inventory — no longer visible from any room
	_, ok = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.b")
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
	_, ok = game.GetItem("room.tutorial.lock_hall", "item.tutorial.old_lantern", "player.a")
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

func TestTagInstance_Params(t *testing.T) {
	w := New()
	_ = w
	item := Item{
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.foo"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
	}
	params, ok := item.tagParams("tag.exit")
	if !ok {
		t.Fatal("expected to find tag.exit params")
	}
	if params["direction"] != "north" {
		t.Fatalf("direction = %q, want north", params["direction"])
	}
	if params["target"] != "room.foo" {
		t.Fatalf("target = %q, want room.foo", params["target"])
	}
	_, ok = item.tagParams("tag.carryable")
	if !ok {
		t.Fatal("expected to find tag.carryable params")
	}
	_, ok = item.tagParams("tag.nonexistent")
	if ok {
		t.Fatal("unexpected tag.nonexistent params")
	}
}

func TestExitTag_behaviorConsistency(t *testing.T) {
	// 验证 TagInstance 到 Exit 提取与旧行为一致
	w := New()
	_ = w
	item := Item{
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.test_target"}},
		},
	}
	w.items = map[ItemID]Item{"item.test_exit": item}

	exit, ok := w.itemExit("item.test_exit")
	if !ok {
		t.Fatal("expected exit to be found")
	}
	if exit.Direction != "north" {
		t.Fatalf("direction = %q, want north", exit.Direction)
	}
	if exit.TargetRoomID != "room.test_target" {
		t.Fatalf("target = %q, want room.test_target", exit.TargetRoomID)
	}
}

func TestCarryableTag_behaviorConsistency(t *testing.T) {
	w := New()
	_ = w
	w.items = map[ItemID]Item{
		"item.carryable": {
			Tags: []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		},
		"item.not_carryable": {
			Tags: nil,
		},
	}
	if !w.itemIsCarryable("item.carryable") {
		t.Fatal("expected carryable item to be carryable")
	}
	if w.itemIsCarryable("item.not_carryable") {
		t.Fatal("expected non-carryable item to not be carryable")
	}
}

func TestLightableTag_definedAndAccessible(t *testing.T) {
	// 从 New() 创建的教程世界验证旧油灯有 lightable
	w := New()
	item, ok := w.items["item.tutorial.old_lantern"]
	if !ok {
		t.Fatal("expected old lantern in tutorial world")
	}
	params, ok := item.tagParams("tag.lightable")
	if !ok {
		t.Fatal("expected old lantern to have tag.lightable")
	}
	_ = params

	// 从编译内容创建的世界也验证
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	w2 := NewFromSnapshot(compiled.Server, compiled.Client)
	item2, ok := w2.items["item.tutorial.old_lantern"]
	if !ok {
		t.Fatal("expected old lantern in compiled world")
	}
	_, ok = item2.tagParams("tag.lightable")
	if !ok {
		t.Fatal("expected old lantern to have tag.lightable in compiled world")
	}
}

func TestContainerTag_registeredAndUsable(t *testing.T) {
	w := New()
	def, ok := w.TagDefinition("tag.container")
	if !ok {
		t.Fatal("expected tag.container to be registered")
	}
	var hasCapacity bool
	for _, f := range def.Fields {
		if f.Name == "capacity" {
			hasCapacity = true
			if f.Type != TagFieldInt {
				t.Fatalf("capacity field type = %q, want int", f.Type)
			}
			defCap, ok := f.Default.(int)
			if !ok || defCap != 1 {
				t.Fatalf("capacity default = %v (type %T), want 1", f.Default, f.Default)
			}
		}
	}
	if !hasCapacity {
		t.Fatal("container tag missing capacity field")
	}

	// 用 tag.container 创建物品并读取参数
	w.items["item.test_backpack"] = Item{
		Name: "背包",
		Tags: []TagInstance{
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		},
	}
	params, ok := w.items["item.test_backpack"].tagParams("tag.container")
	if !ok {
		t.Fatal("expected backpack to have tag.container")
	}
	capacity, ok := params["capacity"].(int)
	if !ok || capacity != 5 {
		t.Fatalf("capacity = %v (type %T), want 5", params["capacity"], params["capacity"])
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
	params, ok := w.items["item.test.box"].tagParams("tag.container")
	if !ok {
		t.Fatal("expected box to have tag.container in world")
	}
	capacity, ok := params["capacity"].(int)
	if !ok || capacity != 3 {
		t.Fatalf("capacity = %v, want 3", params["capacity"])
	}
}

func TestContainer_OpenClose(t *testing.T) {
	w := New()
	w.items = map[ItemID]Item{
		"item.box": {
			Name: "箱子",
			Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 2}}},
		},
	}
	w.itemLocations["item.box"] = RoomItemLocation{RoomID: w.startRoom}
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
	w.items["item.rock"] = Item{Name: "石头"}
	if w.OpenContainer("item.rock") {
		t.Fatal("non-container should not open")
	}
}

func TestContainer_PutAndGet(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.players[playerID] = PlayerEntity{RoomID: w.startRoom}
	// 准备容器
	w.items["item.box"] = Item{
		Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 2}}},
	}
	w.itemLocations["item.box"] = RoomItemLocation{RoomID: w.startRoom}
	w.OpenContainer("item.box")
	// 准备物品（在玩家背包中）
	w.items["item.apple"] = Item{Name: "苹果"}
	w.itemLocations["item.apple"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
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
	inventory := w.itemsInContainer(PlayerContainerID(playerID))
	if len(inventory) != 1 || inventory[0] != "item.apple" {
		t.Fatalf("inventory = %v, want [item.apple]", inventory)
	}
}

func TestContainer_CapacityLimit(t *testing.T) {
	w := New()
	playerID := PlayerID("test")
	w.players[playerID] = PlayerEntity{RoomID: w.startRoom}
	w.items["item.box"] = Item{
		Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 1}}},
	}
	w.itemLocations["item.box"] = RoomItemLocation{RoomID: w.startRoom}
	w.OpenContainer("item.box")
	w.items["item.a"] = Item{Name: "A"}
	w.items["item.b"] = Item{Name: "B"}
	w.itemLocations["item.a"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
	w.itemLocations["item.b"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
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
	w.players[playerID] = PlayerEntity{RoomID: w.startRoom}
	// 便携容器（收纳袋）— carryable + container
	w.items["item.bag"] = Item{
		Name: "收纳袋",
		Tags: []TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		},
	}
	w.itemLocations["item.bag"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
	// 另一个 carryable 容器 — 应该不能放入收纳袋
	w.items["item.pouch"] = Item{
		Name: "小袋",
		Tags: []TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 3}},
		},
	}
	w.itemLocations["item.pouch"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
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
	w.players[playerID] = PlayerEntity{RoomID: w.startRoom}
	w.items["item.box"] = Item{
		Name: "箱子",
		Tags: []TagInstance{{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}}},
	}
	w.itemLocations["item.box"] = RoomItemLocation{RoomID: w.startRoom}
	// 不打开
	w.items["item.apple"] = Item{Name: "苹果"}
	w.itemLocations["item.apple"] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
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
