package world

func (w *World) GetItem(roomID RoomID, name string, playerID PlayerID) (ItemID, bool) {
	for _, itemID := range w.itemsInRoom(roomID) {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		if item.Name != name {
			continue
		}

		w.itemLocations[itemID] = InventoryItemLocation{PlayerID: playerID}
		return itemID, true
	}

	return "", false
}

func (w *World) DropItem(roomID RoomID, itemID ItemID) {
	w.itemLocations[itemID] = RoomItemLocation{RoomID: roomID}
}

func (w *World) DropItemByName(roomID RoomID, name string, playerID PlayerID) bool {
	for _, itemID := range w.itemsInInventory(playerID) {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		if item.Name != name {
			continue
		}

		w.DropItem(roomID, itemID)
		return true
	}

	return false
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
