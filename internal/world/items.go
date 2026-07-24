package world

import (
	"slices"
	"strings"
)

func (w *World) GetItem(roomID, targetItemID, playerID EntityID) (EntityID, bool, bool) {
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if eid != targetItemID || w.store.Item(eid) == nil {
			continue
		}
		if !w.store.Tag(eid, "tag.carryable") {
			continue
		}
		volOk, _ := w.CanAddItem(playerID, eid)
		if !volOk {
			return eid, false, false
		}
		w.store.RemoveFromRoom(eid)
		cid := PlayerContainerID(playerID)
		w.containerContents[cid] = append(w.containerContents[cid], eid)
		return eid, true, true
	}
	return "", false, true
}

func (w *World) DropItem(roomID, itemID EntityID) bool {
	ed := w.store.Exit(itemID)
	if ed != nil && ed.Direction != "" {
		for _, eid := range w.store.EntitiesInRoom(roomID) {
			if eid == itemID {
				continue
			}
			if other := w.store.Exit(eid); other != nil && other.Direction == ed.Direction {
				return false
			}
		}
	}
	w.removeFromContainer(itemID)
	w.store.PlaceInRoom(itemID, roomID)
	return true
}

func (w *World) DropInventoryItem(roomID, targetItemID, playerID EntityID) (EntityID, bool) {
	for _, eid := range w.containerContents[PlayerContainerID(playerID)] {
		if eid != targetItemID {
			continue
		}
		if w.DropItem(roomID, eid) {
			return eid, true
		}
		return "", false
	}
	return "", false
}

func (w *World) ExamineItem(roomID, targetItemID, playerID EntityID) (ItemObservation, bool) {
	if !w.RoomIsLit(roomID, playerID) {
		return ItemObservation{}, false
	}
	for _, eid := range w.visibleItemIDs(roomID, playerID) {
		if eid != targetItemID {
			continue
		}
		ent := w.store.Get(eid)
		if ent == nil {
			return ItemObservation{}, false
		}
		rootTags := observableTagDescriptions(ent.Tags, w)
		partTags := observablePartTagDescriptions(ent.Parts, w)
		return ItemObservation{
			Item:           eid,
			NameKey:        ent.NameKey,
			DescriptionKey: ent.DescriptionKey,
			Name:           ent.Name,
			Description:    ent.Description,
			Tags:           rootTags,
			PartTags:       partTags,
		}, true
	}
	return ItemObservation{}, false
}

func (w *World) ResolveRoomItemPhrase(roomID EntityID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.carryableItemsInRoom(roomID), phrase)
}

func (w *World) ResolveInventoryItemPhrase(playerID EntityID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.containerContents[PlayerContainerID(playerID)], phrase)
}

func (w *World) ResolveVisibleItemPhrase(roomID, playerID EntityID, phrase string) ItemResolution {
	return w.resolveItemPhrase(w.visibleItemIDs(roomID, playerID), phrase)
}

func (w *World) resolveItemPhrase(entityIDs []EntityID, phrase string) ItemResolution {
	matches := make([]EntityID, 0, 1)
	for _, eid := range entityIDs {
		ent := w.store.Get(eid)
		if ent == nil {
			continue
		}
		if ent.matchesPhrase(eid, phrase) {
			matches = append(matches, eid)
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

func (e *Entity) matchesPhrase(entityID EntityID, phrase string) bool {
	if phrase == string(entityID) {
		return true
	}
	if phrase == e.Name {
		return true
	}
	normalizedPhrase := normalizeInputName(phrase)
	if normalizedPhrase == "" {
		return false
	}
	if normalizedPhrase == normalizeInputName(e.Name) {
		return true
	}
	if normalizedPhrase == normalizeInputName(e.InnerName) {
		return true
	}
	for _, alias := range e.Aliases {
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

func (w *World) ItemName(itemID EntityID) (string, bool) {
	ent := w.store.Get(itemID)
	if ent == nil {
		return "", false
	}
	return ent.Name, true
}

func (w *World) ItemNameOr(itemID EntityID) string {
	ent := w.store.Get(itemID)
	if ent == nil {
		return "未知"
	}
	return ent.Name
}

func (w *World) ItemNames(itemIDs []EntityID) []string {
	return w.itemNames(itemIDs)
}

func (w *World) Inventory(playerID EntityID) []string {
	return w.itemNames(w.containerContents[PlayerContainerID(playerID)])
}

func (w *World) InventoryItemIDs(playerID EntityID) []EntityID {
	return w.containerContents[PlayerContainerID(playerID)]
}

func (w *World) itemTotalWeight(itemID EntityID) int {
	itemData := w.store.Item(itemID)
	if itemData == nil {
		return 0
	}
	total := itemData.Weight
	for _, contentID := range w.containerContents[ItemContainerID(itemID)] {
		total += w.itemTotalWeight(contentID)
	}
	return total
}

func (w *World) PlayerCurrentWeight(playerID EntityID) int {
	total := 0
	for _, eid := range w.containerContents[PlayerContainerID(playerID)] {
		total += w.itemTotalWeight(eid)
	}
	return total
}

func (w *World) PlayerCurrentVolume(playerID EntityID) int {
	total := 0
	for _, eid := range w.containerContents[PlayerContainerID(playerID)] {
		if itemData := w.store.Item(eid); itemData != nil {
			total += itemData.Volume
		}
	}
	return total
}

func (w *World) CanAddItem(playerID, itemID EntityID) (volumeOk, weightOk bool) {
	playerData := w.store.Player(playerID)
	if playerData == nil {
		return true, true
	}
	itemData := w.store.Item(itemID)
	if itemData == nil {
		return false, false
	}
	newVol := w.PlayerCurrentVolume(playerID) + itemData.Volume
	volumeOk = playerData.MaxVolume <= 0 || newVol <= playerData.MaxVolume
	newWt := w.PlayerCurrentWeight(playerID) + w.itemTotalWeight(itemID)
	weightOk = playerData.MaxWeight <= 0 || newWt <= playerData.MaxWeight
	return
}

func (w *World) IsOverWeight(playerID EntityID) bool {
	playerData := w.store.Player(playerID)
	if playerData == nil {
		return false
	}
	if playerData.MaxWeight <= 0 {
		return false
	}
	return w.PlayerCurrentWeight(playerID) > playerData.MaxWeight
}

func (w *World) PlayerWeightRatio(playerID EntityID) (current, max int) {
	playerData := w.store.Player(playerID)
	if playerData == nil {
		return 0, 0
	}
	return w.PlayerCurrentWeight(playerID), playerData.MaxWeight
}

func (w *World) PlayerVolumeRatio(playerID EntityID) (current, max int) {
	playerData := w.store.Player(playerID)
	if playerData == nil {
		return 0, 0
	}
	return w.PlayerCurrentVolume(playerID), playerData.MaxVolume
}

func (w *World) itemNames(itemIDs []EntityID) []string {
	names := make([]string, 0, len(itemIDs))
	for _, eid := range itemIDs {
		name, ok := w.ItemName(eid)
		if !ok {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (w *World) itemsInRoom(roomID EntityID) []EntityID {
	var result []EntityID
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if w.store.Item(eid) != nil {
			result = append(result, eid)
		}
	}
	return result
}

func (w *World) visibleItemIDs(roomID, playerID EntityID) []EntityID {
	var result []EntityID
	result = append(result, w.itemsInRoom(roomID)...)
	result = append(result, w.containerContents[PlayerContainerID(playerID)]...)
	return result
}

func (w *World) ordinaryItemsInRoom(roomID EntityID) []EntityID {
	var result []EntityID
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if w.store.Item(eid) != nil && w.store.Exit(eid) == nil {
			result = append(result, eid)
		}
	}
	return result
}

func (w *World) carryableItemsInRoom(roomID EntityID) []EntityID {
	var result []EntityID
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if w.store.Item(eid) != nil && w.store.Tag(eid, "tag.carryable") {
			result = append(result, eid)
		}
	}
	return result
}

func (w *World) exitItemIDs(roomID EntityID) []EntityID {
	var result []EntityID
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if w.store.Exit(eid) != nil {
			result = append(result, eid)
		}
	}
	return result
}

func (w *World) itemIsCarryable(itemID EntityID) bool {
	return w.store.Item(itemID) != nil && w.store.Tag(itemID, "tag.carryable")
}

func (w *World) removeFromContainer(entityID EntityID) {
	for cid, entities := range w.containerContents {
		for i, eid := range entities {
			if eid == entityID {
				w.containerContents[cid] = append(entities[:i], entities[i+1:]...)
				return
			}
		}
	}
}

func observableTagDescriptions(tags []TagInstance, w *World) []string {
	var result []string
	for _, inst := range tags {
		def, ok := w.TagDefinition(inst.DefinitionID)
		if ok && def.Observable {
			result = append(result, def.Description)
		}
	}
	return result
}

func observablePartTagDescriptions(parts map[string]ItemPart, w *World) map[string][]string {
	if len(parts) == 0 {
		return nil
	}
	result := make(map[string][]string)
	for partID, part := range parts {
		for _, inst := range part.Tags {
			def, ok := w.TagDefinition(inst.DefinitionID)
			if ok && def.Observable {
				result[partID] = append(result[partID], def.Description)
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
