package world

import (
	"slices"
	"strings"
)

func (w *World) GetItem(roomID RoomID, targetItemID ItemID, playerID PlayerID) (ItemID, bool) {
	for _, itemID := range w.carryableItemsInRoom(roomID) {
		if itemID != targetItemID {
			continue
		}

		w.itemLocations[itemID] = InventoryItemLocation{PlayerID: playerID}
		return itemID, true
	}

	return "", false
}

func (w *World) DropItem(roomID RoomID, itemID ItemID) bool {
	if exit, ok := w.itemExit(itemID); ok && exit.Direction != "" {
		for _, existingID := range w.exitItemIDs(roomID) {
			if existingID == itemID {
				continue
			}
			existing, exists := w.itemExit(existingID)
			if exists && existing.Direction == exit.Direction {
				return false
			}
		}
	}
	w.itemLocations[itemID] = RoomItemLocation{RoomID: roomID}
	return true
}

func (w *World) DropInventoryItem(roomID RoomID, targetItemID ItemID, playerID PlayerID) (ItemID, bool) {
	for _, itemID := range w.itemsInInventory(playerID) {
		if itemID != targetItemID {
			continue
		}

		if w.DropItem(roomID, itemID) {
			return itemID, true
		}
		return "", false
	}

	return "", false
}

func (w *World) ExamineItem(roomID RoomID, targetItemID ItemID, playerID PlayerID) (ItemObservation, bool) {
	for _, itemID := range w.visibleItemIDs(roomID, playerID) {
		if itemID != targetItemID {
			continue
		}
		item, ok := w.items[itemID]
		if !ok {
			return ItemObservation{}, false
		}
		return ItemObservation{
			Item:           itemID,
			NameKey:        item.NameKey,
			DescriptionKey: item.DescriptionKey,
			Name:           item.Name,
			Description:    item.Description,
		}, true
	}
	return ItemObservation{}, false
}

func (w *World) ResolveRoomItemPhrase(roomID RoomID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.carryableItemsInRoom(roomID), phrase)
}

func (w *World) ResolveInventoryItemPhrase(playerID PlayerID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.itemsInInventory(playerID), phrase)
}

func (w *World) ResolveVisibleItemPhrase(roomID RoomID, playerID PlayerID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.visibleItemIDs(roomID, playerID), phrase)
}

func (w *World) resolveItemPhrase(itemIDs []ItemID, phrase string) ItemResolution {
	matches := make([]ItemID, 0, 1)
	for _, itemID := range itemIDs {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		if item.matchesPhrase(itemID, phrase) {
			matches = append(matches, itemID)
		}
	}
	slices.Sort(matches)
	if len(matches) == 0 {
		return ItemResolution{}
	}
	if len(matches) > 1 {
		return ItemResolution{AmbiguousItemIDs: matches}
	}
	return ItemResolution{ItemID: matches[0], Found: true}
}

func (i Item) matchesPhrase(itemID ItemID, phrase string) bool {
	if phrase == string(itemID) {
		return true
	}
	if phrase == i.Name {
		return true
	}
	normalizedPhrase := normalizeInputName(phrase)
	if normalizedPhrase == "" {
		return false
	}
	if normalizedPhrase == normalizeInputName(i.InnerName) {
		return true
	}
	for _, alias := range i.Aliases {
		if normalizedPhrase == normalizeInputName(alias) {
			return true
		}
	}
	return false
}

func normalizeInputName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, " ", "")
	return name
}

func (w *World) ItemName(itemID ItemID) (string, bool) {
	item, ok := w.items[itemID]
	if !ok {
		return "", false
	}
	return item.Name, true
}

func (w *World) ItemNames(itemIDs []ItemID) []string {
	return w.itemNames(itemIDs)
}

func (w *World) Inventory(playerID PlayerID) []string {
	return w.itemNames(w.itemsInInventory(playerID))
}

func (w *World) InventoryItemIDs(playerID PlayerID) []ItemID {
	return w.itemsInInventory(playerID)
}

func (w *World) itemNames(itemIDs []ItemID) []string {
	names := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		name, ok := w.ItemName(itemID)
		if !ok {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (w *World) itemsInRoom(roomID RoomID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for itemID, location := range w.itemLocations {
		roomLocation, ok := location.(RoomItemLocation)
		if !ok {
			continue
		}
		if roomLocation.RoomID == roomID {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) itemsInInventory(playerID PlayerID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for itemID, location := range w.itemLocations {
		inventoryLocation, ok := location.(InventoryItemLocation)
		if !ok {
			continue
		}
		if inventoryLocation.PlayerID == playerID {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) visibleItemIDs(roomID RoomID, playerID PlayerID) []ItemID {
	itemIDs := w.itemsInRoom(roomID)
	itemIDs = append(itemIDs, w.itemsInInventory(playerID)...)
	return itemIDs
}

func (w *World) ordinaryItemsInRoom(roomID RoomID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for _, itemID := range w.itemsInRoom(roomID) {
		if _, isExit := w.itemExit(itemID); !isExit {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) carryableItemsInRoom(roomID RoomID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for _, itemID := range w.itemsInRoom(roomID) {
		if w.itemIsCarryable(itemID) {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) exitItemIDs(roomID RoomID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for _, itemID := range w.itemsInRoom(roomID) {
		if _, isExit := w.itemExit(itemID); isExit {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) itemExit(itemID ItemID) (Exit, bool) {
	item, ok := w.items[itemID]
	if !ok {
		return Exit{}, false
	}
	for _, tag := range item.Tags {
		if tag.Exit != nil {
			return *tag.Exit, true
		}
	}
	return Exit{}, false
}

func (w *World) itemIsCarryable(itemID ItemID) bool {
	item, ok := w.items[itemID]
	if !ok {
		return false
	}
	for _, tag := range item.Tags {
		if tag.Carryable {
			return true
		}
	}
	return false
}
