package world

func (w *World) LightItem(itemID EntityID) {
	w.litItems[itemID] = true
}

func (w *World) ExtinguishItem(itemID EntityID) {
	delete(w.litItems, itemID)
}

func (w *World) IsItemLit(itemID EntityID) bool {
	return w.litItems[itemID]
}

func (w *World) RoomIsLit(roomID, playerID EntityID) bool {
	rd := w.store.Room(roomID)
	if rd == nil {
		return false
	}
	if !rd.Dark {
		return true
	}
	for _, eid := range w.containerContents[PlayerContainerID(playerID)] {
		if w.IsItemLit(eid) {
			return true
		}
	}
	for _, eid := range w.itemsInRoom(roomID) {
		if w.IsItemLit(eid) {
			return true
		}
	}
	return false
}
