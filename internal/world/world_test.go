package world

import (
	"PMud/internal/content"
	"slices"
	"testing"
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
	itemID, ok := game.GetItem(game.StartRoom(), "旧油灯", playerID)
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

func TestWorld_ItemMovesBetweenRoomAndInventory(t *testing.T) {
	// Given
	game := New()
	playerID := PlayerID("player.local")
	startRoom := game.StartRoom()

	// When
	itemID, ok := game.GetItem(startRoom, "旧油灯", playerID)

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
	ok = game.DropItemByName(startRoom, "旧油灯", playerID)

	// Then
	if !ok {
		t.Fatal("expected to drop old lantern")
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
