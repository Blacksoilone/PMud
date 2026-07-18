package world

import (
	"slices"
	"testing"

	"PMud/internal/content"
)

func TestWorld_NewFromSnapshotPreservesTutorialBehavior(t *testing.T) {
	// Given
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	game := NewFromSnapshot(compiled.Server, compiled.Client)
	playerID := PlayerID("player.local")

	// When
	observation, ok := game.Look(game.StartRoom())

	// Then
	if !ok {
		t.Fatal("expected start room to exist")
	}
	if observation.Name != "练习场入口" {
		t.Fatalf("expected start room name, got %q", observation.Name)
	}
	if !slices.Contains(observation.Items, "旧油灯") {
		t.Fatalf("expected old lantern in start room, got %v", observation.Items)
	}
	nextRoom, ok := game.Move(game.StartRoom(), "north")
	if !ok {
		t.Fatal("expected north movement to work")
	}
	if nextRoom != "room.tutorial.yard" {
		t.Fatalf("expected yard room, got %q", nextRoom)
	}
	itemID, ok := game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)
	if !ok {
		t.Fatal("expected to get old lantern")
	}
	if itemID != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", itemID)
	}
	if !slices.Contains(game.Inventory(playerID), "旧油灯") {
		t.Fatalf("expected old lantern in inventory, got %v", game.Inventory(playerID))
	}
}

func TestWorldMoveUsesExitItems(t *testing.T) {
	// Given
	game := New()
	startRoom := game.StartRoom()
	game.rooms[startRoom] = Room{
		NameKey:        game.rooms[startRoom].NameKey,
		DescriptionKey: game.rooms[startRoom].DescriptionKey,
		Name:           game.rooms[startRoom].Name,
		Description:    game.rooms[startRoom].Description,
	}
	game.items["item.tutorial.north"] = Item{
		Name:      "北方",
		InnerName: "north",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.yard"}},
		},
	}
	game.itemLocations["item.tutorial.north"] = RoomItemLocation{RoomID: startRoom}

	// When
	nextRoom, moved := game.Move(startRoom, "north")

	// Then
	if !moved || nextRoom != "room.tutorial.yard" {
		t.Fatalf("Move north = %q, %v; want yard, true", nextRoom, moved)
	}
}

func TestWorldMoveUsesNamedExitWithoutDirection(t *testing.T) {
	// Given
	game := New()
	startRoom := game.StartRoom()
	game.rooms[startRoom] = Room{Name: game.rooms[startRoom].Name}
	game.items["item.tutorial.portal"] = Item{
		Name:      "传送门",
		InnerName: "portal",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"target": "room.tutorial.yard"}},
		},
	}
	game.itemLocations["item.tutorial.portal"] = RoomItemLocation{RoomID: startRoom}

	// When
	nextRoom, moved := game.Move(startRoom, "传送门")

	// Then
	if !moved || nextRoom != "room.tutorial.yard" {
		t.Fatalf("Move portal = %q, %v; want yard, true", nextRoom, moved)
	}
}

func TestWorldLookSeparatesExitsFromOrdinaryItems(t *testing.T) {
	// Given
	game := New()
	startRoom := game.StartRoom()
	game.items["item.tutorial.north"] = Item{
		Name:      "北方",
		InnerName: "north",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.yard"}},
		},
	}
	game.itemLocations["item.tutorial.north"] = RoomItemLocation{RoomID: startRoom}

	// When
	observation, ok := game.Look(startRoom)
	_, gotExit := game.GetItem(startRoom, "item.tutorial.north", "player.local")

	// Then
	if !ok {
		t.Fatal("expected room observation")
	}
	if !slices.Contains(observation.Exits, "north") {
		t.Fatalf("exits = %v, want north", observation.Exits)
	}
	if got := observation.Neighbors["north"]; got != "room.tutorial.yard" {
		t.Fatalf("north neighbor = %q, want room.tutorial.yard", got)
	}
	if slices.Contains(observation.Items, "北方") {
		t.Fatalf("ordinary items include exit: %v", observation.Items)
	}
	if gotExit {
		t.Fatal("expected exit item not to be gettable")
	}
}

func TestWorld_ItemMovesBetweenRoomAndInventory(t *testing.T) {
	// Given
	game := New()
	playerID := PlayerID("player.local")
	startRoom := game.StartRoom()

	// When
	itemID, ok := game.GetItem(startRoom, "item.tutorial.old_lantern", playerID)

	// Then
	if !ok {
		t.Fatal("expected to get old lantern")
	}
	observation, ok := game.Look(startRoom)
	if !ok {
		t.Fatal("expected start room to exist")
	}
	if slices.Contains(observation.Items, "旧油灯") {
		t.Fatal("expected old lantern to leave the room after get")
	}
	if !slices.Contains(game.Inventory(playerID), "旧油灯") {
		t.Fatal("expected old lantern to enter player inventory after get")
	}

	// When
	droppedItemID, ok := game.DropInventoryItem(startRoom, "item.tutorial.old_lantern", playerID)

	// Then
	if !ok {
		t.Fatal("expected to drop old lantern")
	}
	if droppedItemID != itemID {
		t.Fatalf("expected dropped item %q, got %q", itemID, droppedItemID)
	}
	observation, ok = game.Look(startRoom)
	if !ok {
		t.Fatal("expected start room to exist")
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

	observation, ok := game.ExamineItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)

	if !ok {
		t.Fatal("expected to examine old lantern in current room")
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
	game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)

	observation, ok := game.ExamineItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)

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
	// Given
	game := New()

	// When
	resolution := game.ResolveRoomItemPhrase(game.StartRoom(), "旧油灯")

	// Then
	if !resolution.Found {
		t.Fatal("expected old lantern to resolve")
	}
	if resolution.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("item id = %q, want old lantern", resolution.ItemID)
	}
}

func TestWorldResolveRoomItemPhrase_matchesInnerNameSeparatorsCaseInsensitively(t *testing.T) {
	// Given
	game := New()

	tests := []string{"oldlantern", "old-lantern", "old_lantern", "OLD-LANTERN"}
	for _, phrase := range tests {
		t.Run(phrase, func(t *testing.T) {
			// When
			resolution := game.ResolveRoomItemPhrase(game.StartRoom(), phrase)

			// Then
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
	// Given
	game := New()

	tests := []string{"jiuyoudeng", "jiu-youdeng", "jiu_youdeng", "JIU_YOUDENG"}
	for _, phrase := range tests {
		t.Run(phrase, func(t *testing.T) {
			// When
			resolution := game.ResolveRoomItemPhrase(game.StartRoom(), phrase)

			// Then
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
	// Given
	game := New()

	// When
	resolution := game.ResolveRoomItemPhrase("room.tutorial.yard", "lianximujian")

	// Then
	if !resolution.Found {
		t.Fatal("expected practice sword to resolve")
	}
	if resolution.ItemID != "item.tutorial.practice_sword" {
		t.Fatalf("item id = %q, want practice sword", resolution.ItemID)
	}
}

func TestWorldResolveInventoryItemPhrase_matchesOnlyInventory(t *testing.T) {
	// Given
	game := New()
	playerID := PlayerID("player.local")
	game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)

	// When
	inventoryResolution := game.ResolveInventoryItemPhrase(playerID, "旧油灯")
	roomResolution := game.ResolveRoomItemPhrase(game.StartRoom(), "旧油灯")

	// Then
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
	// Given
	game := New()
	playerID := PlayerID("player.local")
	game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", playerID)

	// When
	inventoryResolution := game.ResolveVisibleItemPhrase(game.StartRoom(), playerID, "旧油灯")
	yardResolution := game.ResolveVisibleItemPhrase("room.tutorial.yard", playerID, "练习木剑")

	// Then
	if !inventoryResolution.Found || inventoryResolution.ItemID != "item.tutorial.old_lantern" {
		t.Fatalf("inventory visible resolution = %#v, want old lantern", inventoryResolution)
	}
	if !yardResolution.Found || yardResolution.ItemID != "item.tutorial.practice_sword" {
		t.Fatalf("yard visible resolution = %#v, want practice sword", yardResolution)
	}
}

func TestWorldResolveRoomItemPhrase_reportsAmbiguityOnlyWithinRoom(t *testing.T) {
	// Given
	game := New()
	game.items["item.tutorial.second_lantern"] = Item{
		NameKey:        "item.tutorial.second_lantern.name",
		DescriptionKey: "item.tutorial.second_lantern.description",
		Name:           "旧油灯",
		InnerName:      "old lantern",
		Description:    "另一盏旧油灯。",
		Aliases:        []string{"jiuyoudeng"},
		Tags:           []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
	}
	game.itemLocations["item.tutorial.second_lantern"] = RoomItemLocation{RoomID: game.StartRoom()}
	game.items["item.tutorial.distant_lantern"] = Item{
		NameKey:        "item.tutorial.distant_lantern.name",
		DescriptionKey: "item.tutorial.distant_lantern.description",
		Name:           "旧油灯",
		InnerName:      "old lantern",
		Description:    "远处的旧油灯。",
		Aliases:        []string{"jiuyoudeng"},
		Tags:           []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
	}
	game.itemLocations["item.tutorial.distant_lantern"] = RoomItemLocation{RoomID: "room.tutorial.yard"}

	// When
	startResolution := game.ResolveRoomItemPhrase(game.StartRoom(), "旧油灯")
	yardResolution := game.ResolveRoomItemPhrase("room.tutorial.yard", "旧油灯")

	// Then
	wantAmbiguous := []ItemID{"item.tutorial.old_lantern", "item.tutorial.second_lantern"}
	if startResolution.Found {
		t.Fatalf("expected ambiguous start room resolution, got found %#v", startResolution)
	}
	if !slices.Equal(startResolution.AmbiguousItemIDs, wantAmbiguous) {
		t.Fatalf("ambiguous ids = %#v, want %#v", startResolution.AmbiguousItemIDs, wantAmbiguous)
	}
	if !yardResolution.Found || yardResolution.ItemID != "item.tutorial.distant_lantern" {
		t.Fatalf("yard resolution = %#v, want distant lantern only", yardResolution)
	}
}

func TestWorldGetItemAllowsCarryableExit(t *testing.T) {
	// Given
	game := New()
	startRoom := game.StartRoom()
	game.items["item.tutorial.portal"] = Item{
		Name:      "传送门",
		InnerName: "portal",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"target": "room.tutorial.yard"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
	}
	game.itemLocations["item.tutorial.portal"] = RoomItemLocation{RoomID: startRoom}

	// When
	itemID, got := game.GetItem(startRoom, "item.tutorial.portal", "player.local")

	// Then
	if !got || itemID != "item.tutorial.portal" {
		t.Fatalf("GetItem carryable exit = %q, %v", itemID, got)
	}
}

func TestWorldDropInventoryItemRejectsDuplicateExitDirection(t *testing.T) {
	// Given
	game := New()
	playerID := PlayerID("player.local")
	startRoom := game.StartRoom()
	game.items["item.tutorial.moving_north"] = Item{
		Name:      "另一个北方",
		InnerName: "north",
		Tags: []TagInstance{
			{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.yard"}},
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
		},
	}
	game.itemLocations["item.tutorial.moving_north"] = InventoryItemLocation{PlayerID: playerID}

	// When
	itemID, dropped := game.DropInventoryItem(startRoom, "item.tutorial.moving_north", playerID)

	// Then
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

	// A picks up lantern
	_, ok = game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", "player.a")
	if !ok {
		t.Fatal("A should get the lantern")
	}

	// B should not see lantern in room
	roomB, _ := game.Look(game.StartRoom())
	if slices.Contains(roomB.Items, "旧油灯") {
		t.Fatal("B should not see lantern in room after A picked it up")
	}

	// B should not be able to get lantern
	_, ok = game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", "player.b")
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

func TestLockedExit_blocksMoveWithoutKey(t *testing.T) {
	game := New()
	game.EnterWorld("player.a")

	_, _, reason := game.MovePlayer("player.a", "north")
	if reason != "locked" {
		t.Fatalf("expected locked reason, got %q", reason)
	}
	room, _ := game.PlayerRoom("player.a")
	if room != game.StartRoom() {
		t.Fatalf("expected to stay in start room, got %q", room)
	}
}

func TestLockedExit_allowsMoveWithKey(t *testing.T) {
	game := New()
	game.EnterWorld("player.a")

	game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", "player.a")
	nextRoom, ok, reason := game.MovePlayer("player.a", "north")
	if !ok {
		t.Fatalf("expected success with key, got reason=%q", reason)
	}
	if nextRoom != "room.tutorial.yard" {
		t.Fatalf("expected yard, got %q", nextRoom)
	}
}

func TestNonLockedExit_allowsMoveWithoutKey(t *testing.T) {
	game := New()
	game.EnterWorld("player.a")

	nextRoom, ok, reason := game.MovePlayer("player.a", "northeast")
	if !ok {
		t.Fatalf("expected northeast to be accessible, got reason=%q", reason)
	}
	if nextRoom != "room.tutorial.shed" {
		t.Fatalf("expected shed, got %q", nextRoom)
	}
}

func TestTwoPlayersMoveIndependently(t *testing.T) {
	game := New()

	game.EnterWorld("player.a")
	game.EnterWorld("player.b")

	// A picks up old lantern (key to the locked north door) then moves north
	_, ok := game.GetItem(game.StartRoom(), "item.tutorial.old_lantern", "player.a")
	if !ok {
		t.Fatal("A should be able to get old lantern")
	}
	roomA, ok, _ := game.MovePlayer("player.a", "north")
	if !ok {
		t.Fatal("A should be able to move north with key")
	}
	if roomA != "room.tutorial.yard" {
		t.Fatalf("A should be in yard, got %q", roomA)
	}

	// B should still be in start
	roomB, ok := game.PlayerRoom("player.b")
	if !ok {
		t.Fatal("player B should exist")
	}
	if roomB != game.StartRoom() {
		t.Fatalf("B should still be in start room, got %q", roomB)
	}

	// Only A in yard
	yardPlayers := game.PlayersInRoom("room.tutorial.yard")
	if !slices.Contains(yardPlayers, PlayerID("player.a")) {
		t.Fatal("A should be in yard")
	}
	if slices.Contains(yardPlayers, PlayerID("player.b")) {
		t.Fatal("B should not be in yard")
	}

	// Only B in start
	startPlayers := game.PlayersInRoom(game.StartRoom())
	if !slices.Contains(startPlayers, PlayerID("player.b")) {
		t.Fatal("B should be in start")
	}
	if slices.Contains(startPlayers, PlayerID("player.a")) {
		t.Fatal("A should not be in start")
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
