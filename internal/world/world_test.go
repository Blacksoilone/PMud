package world

import (
	"slices"
	"testing"
)

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
