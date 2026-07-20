package world

func (w *World) LightItem(itemID ItemID) {
	w.litItems[itemID] = true
}

func (w *World) ExtinguishItem(itemID ItemID) {
	delete(w.litItems, itemID)
}

func (w *World) IsItemLit(itemID ItemID) bool {
	return w.litItems[itemID]
}

// RoomIsLit 检查房间是否有光照。
// 自然采光的房间（Dark=false）永远有光；
// 黑暗房间（Dark=true）需要玩家或房间内有一个点亮的光源。
func (w *World) RoomIsLit(roomID RoomID, playerID PlayerID) bool {
	room, ok := w.rooms[roomID]
	if !ok {
		return false
	}
	if !room.Dark {
		return true
	}
	// 检查玩家背包中是否有点亮的光源
	for _, itemID := range w.itemsInContainer(PlayerContainerID(playerID)) {
		if w.IsItemLit(itemID) {
			return true
		}
	}
	// 检查房间中是否有点亮的光源
	for _, itemID := range w.itemsInRoom(roomID) {
		if w.IsItemLit(itemID) {
			return true
		}
	}
	return false
}
