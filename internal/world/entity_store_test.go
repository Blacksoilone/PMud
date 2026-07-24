package world

import (
	"testing"
)

// ---------------------------------------------------------------------------
// 辅助构造
// ---------------------------------------------------------------------------

func testEntityRoom(id, name string, dark bool, tags ...TagInstance) *Entity {
	return &Entity{
		ID:   id,
		Name: name,
		Tags: tags,
		Room: &RoomData{Dark: dark},
	}
}

func testEntityItem(id, name string, weight, volume int, tags ...TagInstance) *Entity {
	return &Entity{
		ID:   id,
		Name: name,
		Tags: tags,
		Item: &ItemData{Weight: weight, Volume: volume},
	}
}

func testEntityExit(id, direction, target string, tags ...TagInstance) *Entity {
	return &Entity{
		ID:   id,
		Name: direction + " exit",
		Tags: tags,
		Exit: &ExitData{Direction: direction, TargetRoomID: target},
	}
}

func testEntityPlayer(id string) *Entity {
	return &Entity{
		ID:     id,
		Name:   "player",
		Player: &PlayerData{MaxWeight: 20, MaxVolume: 10},
	}
}

// ---------------------------------------------------------------------------
// 基本 CRUD
// ---------------------------------------------------------------------------

func TestEntityStore_AddAndGet(t *testing.T) {
	s := NewEntityStore()
	room := testEntityRoom("room:test", "Test Room", false)
	s.Add(room)

	got := s.Get("room:test")
	if got == nil {
		t.Fatal("expected entity")
	}
	if got.Name != "Test Room" {
		t.Fatalf("name = %q, want %q", got.Name, "Test Room")
	}
	if got.Room == nil || got.Room.Dark {
		t.Fatal("expected non-dark room data")
	}
}

func TestEntityStore_GetMissing(t *testing.T) {
	s := NewEntityStore()
	got := s.Get("room:nonexistent")
	if got != nil {
		t.Fatal("expected nil for missing entity")
	}
}

func TestEntityStore_Has(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	if !s.Has("room:a") {
		t.Fatal("expected Has to be true")
	}
	if s.Has("room:b") {
		t.Fatal("expected Has to be false for missing")
	}
}

func TestEntityStore_AddOverwrites(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "Old", false))
	s.Add(testEntityRoom("room:a", "New", true))

	got := s.Get("room:a")
	if got.Name != "New" {
		t.Fatalf("name = %q, want New", got.Name)
	}
	if got.Room == nil || !got.Room.Dark {
		t.Fatal("expected dark room from overwrite")
	}
}

// ---------------------------------------------------------------------------
// Tag 索引
// ---------------------------------------------------------------------------

func TestEntityStore_FindByTagNoMatch(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	s.Add(testEntityItem("item:b", "B", 1, 1))

	got := s.FindByTag("tag.nonexistent")
	if got != nil {
		t.Fatal("expected nil for non-existent tag")
	}
}

func TestEntityStore_FindByTagSingle(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false, TagInstance{DefinitionID: "tag.haunted"}))
	s.Add(testEntityItem("item:b", "B", 1, 1))

	got := s.FindByTag("tag.haunted")
	if len(got) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(got))
	}
	if got[0].ID != "room:a" {
		t.Fatalf("id = %q, want room:a", got[0].ID)
	}
}

func TestEntityStore_FindByTagMultipleTypes(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false, TagInstance{DefinitionID: "tag.flammable"}))
	s.Add(testEntityItem("item:b", "Torch", 1, 1, TagInstance{DefinitionID: "tag.flammable"}))
	s.Add(testEntityItem("item:c", "Rock", 2, 2))

	got := s.FindByTag("tag.flammable")
	if len(got) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(got))
	}
	ids := map[string]bool{}
	for _, e := range got {
		ids[e.ID] = true
	}
	if !ids["room:a"] {
		t.Fatal("expected room:a in results")
	}
	if !ids["item:b"] {
		t.Fatal("expected item:b in results")
	}
}

// ---------------------------------------------------------------------------
// 位置
// ---------------------------------------------------------------------------

func TestEntityStore_PlaceInRoomAndQuery(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "Room A", false))
	s.Add(testEntityItem("item:t1", "Torch 1", 1, 1, TagInstance{DefinitionID: "tag.flammable"}))
	s.Add(testEntityItem("item:t2", "Torch 2", 1, 1, TagInstance{DefinitionID: "tag.flammable"}))

	s.PlaceInRoom("item:t1", "room:a")
	s.PlaceInRoom("item:t2", "room:a")

	items := s.EntitiesInRoom("room:a")
	if len(items) != 2 {
		t.Fatalf("expected 2 items in room, got %d", len(items))
	}
}

func TestEntityStore_EntitiesInRoom_returnsCopy(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	s.PlaceInRoom("item:x", "room:a")

	// Mutate the returned slice
	items := s.EntitiesInRoom("room:a")
	items = append(items, "item:y")

	// Original should be unchanged
	items2 := s.EntitiesInRoom("room:a")
	if len(items2) != 1 {
		t.Fatalf("expected 1 item, got %d (was mutated)", len(items2))
	}
}

func TestEntityStore_PlaceInRoom_movesEntity(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	s.Add(testEntityRoom("room:b", "B", false))
	s.Add(testEntityItem("item:x", "X", 1, 1))

	s.PlaceInRoom("item:x", "room:a")
	s.PlaceInRoom("item:x", "room:b")

	if len(s.EntitiesInRoom("room:a")) != 0 {
		t.Fatal("expected item to leave room:a")
	}
	if len(s.EntitiesInRoom("room:b")) != 1 {
		t.Fatal("expected item in room:b")
	}
}

func TestEntityStore_EntitiesInRoom_empty(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))

	items := s.EntitiesInRoom("room:a")
	if items != nil {
		t.Fatal("expected nil for empty room")
	}
}

func TestEntityStore_EntitiesInRoom_unlistedRoom(t *testing.T) {
	s := NewEntityStore()
	items := s.EntitiesInRoom("room:nonexistent")
	if items != nil {
		t.Fatal("expected nil for non-existent room")
	}
}

// ---------------------------------------------------------------------------
// 按标签在房间中查询
// ---------------------------------------------------------------------------

func TestEntityStore_FindByTagInRoom(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "Room A", false, TagInstance{DefinitionID: "tag.haunted"}))
	s.Add(testEntityItem("item:t1", "Torch 1", 1, 1, TagInstance{DefinitionID: "tag.flammable"}))
	s.Add(testEntityItem("item:t2", "Torch 2", 1, 1, TagInstance{DefinitionID: "tag.flammable"}))
	s.Add(testEntityExit("exit:a.north", "north", "room:b", TagInstance{DefinitionID: "tag.exit"}, TagInstance{DefinitionID: "tag.flammable"}))

	s.PlaceInRoom("item:t1", "room:a")
	s.PlaceInRoom("item:t2", "room:a")
	s.PlaceInRoom("exit:a.north", "room:a")

	// 查找 room:a 中所有易燃物（包括房间自身）
	flammable := s.FindByTagInRoom("tag.flammable", "room:a")
	if len(flammable) != 3 {
		t.Fatalf("expected 3 flammable entities in room:a (torch1, torch2, exit), got %d", len(flammable))
	}

	// 查找 haunted（只有房间自身）
	haunted := s.FindByTagInRoom("tag.haunted", "room:a")
	if len(haunted) != 1 {
		t.Fatalf("expected 1 haunted entity (room itself), got %d", len(haunted))
	}
	if haunted[0].ID != "room:a" {
		t.Fatalf("haunted entity = %q, want room:a", haunted[0].ID)
	}
}

func TestEntityStore_FindByTagInRoom_noMatch(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	got := s.FindByTagInRoom("tag.nonexistent", "room:a")
	if got != nil {
		t.Fatal("expected nil")
	}
}

func TestEntityStore_FindByTagInRoom_tagsOnRoomItself(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false, TagInstance{DefinitionID: "tag.dark"}))

	got := s.FindByTagInRoom("tag.dark", "room:a")
	if len(got) != 1 {
		t.Fatalf("expected 1 (the room itself), got %d", len(got))
	}
	if got[0].ID != "room:a" {
		t.Fatalf("id = %q, want room:a", got[0].ID)
	}
}

// ---------------------------------------------------------------------------
// 类型数据访问
// ---------------------------------------------------------------------------

func TestEntityStore_TypedAccessors(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", true))
	s.Add(testEntityItem("item:b", "B", 5, 3))
	s.Add(testEntityExit("exit:c", "north", "room:d"))
	s.Add(testEntityPlayer("player:e"))

	// Room
	rd := s.Room("room:a")
	if rd == nil || !rd.Dark {
		t.Fatal("expected dark room")
	}
	if s.Room("item:b") != nil {
		t.Fatal("expected nil for non-room")
	}

	// Item
	id := s.Item("item:b")
	if id == nil || id.Weight != 5 || id.Volume != 3 {
		t.Fatalf("unexpected item data: %+v", id)
	}
	if s.Item("room:a") != nil {
		t.Fatal("expected nil for non-item")
	}

	// Exit
	ed := s.Exit("exit:c")
	if ed == nil || ed.Direction != "north" || ed.TargetRoomID != "room:d" {
		t.Fatalf("unexpected exit data: %+v", ed)
	}
	if s.Exit("room:a") != nil {
		t.Fatal("expected nil for non-exit")
	}

	// Player
	pd := s.Player("player:e")
	if pd == nil || pd.MaxWeight != 20 || pd.MaxVolume != 10 {
		t.Fatalf("unexpected player data: %+v", pd)
	}
	if s.Player("room:a") != nil {
		t.Fatal("expected nil for non-player")
	}
}

func TestEntityStore_TypedAccessors_missing(t *testing.T) {
	s := NewEntityStore()
	if s.Room("room:x") != nil {
		t.Fatal("expected nil for missing entity")
	}
	if s.Item("room:x") != nil {
		t.Fatal("expected nil for missing entity")
	}
	if s.Exit("room:x") != nil {
		t.Fatal("expected nil for missing entity")
	}
	if s.Player("room:x") != nil {
		t.Fatal("expected nil for missing entity")
	}
}

// ---------------------------------------------------------------------------
// Tag 快速查询
// ---------------------------------------------------------------------------

func TestEntityStore_Tag(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false, TagInstance{DefinitionID: "tag.haunted"}))

	if !s.Tag("room:a", "tag.haunted") {
		t.Fatal("expected room:a to have tag.haunted")
	}
	if s.Tag("room:a", "tag.dark") {
		t.Fatal("expected room:a not to have tag.dark")
	}
	if s.Tag("room:nonexistent", "tag.haunted") {
		t.Fatal("expected false for missing entity")
	}
}

// ---------------------------------------------------------------------------
// 遍历
// ---------------------------------------------------------------------------

func TestEntityStore_Entities(t *testing.T) {
	s := NewEntityStore()
	s.Add(testEntityRoom("room:a", "A", false))
	s.Add(testEntityItem("item:b", "B", 1, 1))

	ids := s.Entities()
	if len(ids) != 2 {
		t.Fatalf("expected 2 entities, got %d", len(ids))
	}
}

func TestEntityStore_Entities_empty(t *testing.T) {
	s := NewEntityStore()
	ids := s.Entities()
	if len(ids) != 0 {
		t.Fatalf("expected 0 entities, got %d", len(ids))
	}
}
