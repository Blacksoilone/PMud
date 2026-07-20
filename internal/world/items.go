package world

import (
	"slices"
	"strings"
)

func (w *World) GetItem(roomID RoomID, targetItemID ItemID, playerID PlayerID) (ItemID, bool, bool) {
	// volumeOk 用于通知调用方（给 handleGet 提示）
	for _, itemID := range w.carryableItemsInRoom(roomID) {
		if itemID != targetItemID {
			continue
		}
		volOk, _ := w.CanAddItem(playerID, itemID)
		if !volOk {
			return itemID, false, false
		}
		w.itemLocations[itemID] = ContainerItemLocation{ContainerID: PlayerContainerID(playerID)}
		return itemID, true, true
	}

	return "", false, true
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
	for _, itemID := range w.itemsInContainer(PlayerContainerID(playerID)) {
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
	if !w.RoomIsLit(roomID, playerID) {
		return ItemObservation{}, false
	}
	for _, itemID := range w.visibleItemIDs(roomID, playerID) {
		if itemID != targetItemID {
			continue
		}
		item, ok := w.items[itemID]
		if !ok {
			return ItemObservation{}, false
		}
		rootTags, partTags := item.ObservableTagDescriptions(w)
		return ItemObservation{
			Item:           itemID,
			NameKey:        item.NameKey,
			DescriptionKey: item.DescriptionKey,
			Name:           item.Name,
			Description:    item.Description,
			Tags:           rootTags,
			PartTags:       partTags,
		}, true
	}
	return ItemObservation{}, false
}

func (w *World) ResolveRoomItemPhrase(roomID RoomID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.carryableItemsInRoom(roomID), phrase)
}

func (w *World) ResolveInventoryItemPhrase(playerID PlayerID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.itemsInContainer(PlayerContainerID(playerID)), phrase)
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
	return w.itemNames(w.itemsInContainer(PlayerContainerID(playerID)))
}

func (w *World) InventoryItemIDs(playerID PlayerID) []ItemID {
	return w.itemsInContainer(PlayerContainerID(playerID))
}

// itemTotalWeight 递归计算物品的总重量，如果物品是容器则包含其内容物
func (w *World) itemTotalWeight(itemID ItemID) int {
	item, ok := w.items[itemID]
	if !ok {
		return 0
	}
	total := item.Weight
	for _, contentID := range w.itemsInContainer(ItemContainerID(itemID)) {
		total += w.itemTotalWeight(contentID)
	}
	return total
}

func (w *World) PlayerCurrentWeight(playerID PlayerID) int {
	total := 0
	for _, itemID := range w.itemsInContainer(PlayerContainerID(playerID)) {
		total += w.itemTotalWeight(itemID)
	}
	return total
}

func (w *World) PlayerCurrentVolume(playerID PlayerID) int {
	total := 0
	for _, itemID := range w.itemsInContainer(PlayerContainerID(playerID)) {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		total += item.Volume
	}
	return total
}

func (w *World) CanAddItem(playerID PlayerID, itemID ItemID) (volumeOk bool, weightOk bool) {
	player, ok := w.players[playerID]
	if !ok {
		return true, true
	}
	item, ok := w.items[itemID]
	if !ok {
		return false, false
	}
	newVol := w.PlayerCurrentVolume(playerID) + item.Volume
	volumeOk = player.MaxVolume <= 0 || newVol <= player.MaxVolume
	newWt := w.PlayerCurrentWeight(playerID) + w.itemTotalWeight(itemID)
	weightOk = player.MaxWeight <= 0 || newWt <= player.MaxWeight
	return
}

func (w *World) IsOverWeight(playerID PlayerID) bool {
	player, ok := w.players[playerID]
	if !ok {
		return false
	}
	if player.MaxWeight <= 0 {
		return false
	}
	return w.PlayerCurrentWeight(playerID) > player.MaxWeight
}

func (w *World) PlayerWeightRatio(playerID PlayerID) (current, max int) {
	player, ok := w.players[playerID]
	if !ok {
		return 0, 0
	}
	return w.PlayerCurrentWeight(playerID), player.MaxWeight
}

func (w *World) PlayerVolumeRatio(playerID PlayerID) (current, max int) {
	player, ok := w.players[playerID]
	if !ok {
		return 0, 0
	}
	return w.PlayerCurrentVolume(playerID), player.MaxVolume
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

func (w *World) itemsInContainer(containerID string) []ItemID {
	itemIDs := make([]ItemID, 0)
	for itemID, location := range w.itemLocations {
		cl, ok := location.(ContainerItemLocation)
		if !ok {
			continue
		}
		if cl.ContainerID == containerID {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) visibleItemIDs(roomID RoomID, playerID PlayerID) []ItemID {
	itemIDs := w.itemsInRoom(roomID)
	itemIDs = append(itemIDs, w.itemsInContainer(PlayerContainerID(playerID))...)
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
	params, ok := item.tagParams("tag.exit")
	if !ok {
		return Exit{}, false
	}
	target, _ := params["target"].(string)
	direction, _ := params["direction"].(string)
	return Exit{Direction: direction, TargetRoomID: RoomID(target)}, true
}

func (w *World) itemIsCarryable(itemID ItemID) bool {
	item, ok := w.items[itemID]
	if !ok {
		return false
	}
	_, ok = item.tagParams("tag.carryable")
	return ok
}
