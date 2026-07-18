package world

func (w *World) EnterWorld(playerID PlayerID) (RoomID, bool) {
	if _, exists := w.players[playerID]; exists {
		return "", false
	}
	w.players[playerID] = PlayerEntity{
		ID:     playerID,
		RoomID: w.startRoom,
	}
	return w.startRoom, true
}

func (w *World) LeaveWorld(playerID PlayerID) {
	delete(w.players, playerID)
}

func (w *World) PlayerCurrentRoom(playerID PlayerID) RoomID {
	if p, ok := w.players[playerID]; ok {
		return p.RoomID
	}
	return ""
}

func (w *World) PlayerRoom(playerID PlayerID) (RoomID, bool) {
	player, ok := w.players[playerID]
	if !ok {
		return "", false
	}
	return player.RoomID, true
}

func (w *World) PlayersInRoom(roomID RoomID) []PlayerID {
	ids := make([]PlayerID, 0)
	for _, p := range w.players {
		if p.RoomID == roomID {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func (w *World) PlayerCount() int {
	return len(w.players)
}

func (w *World) MovePlayer(playerID PlayerID, direction string) (RoomID, bool, string) {
	player, ok := w.players[playerID]
	if !ok {
		return "", false, ""
	}
	for _, itemID := range w.exitItemIDs(player.RoomID) {
		item := w.items[itemID]
		exit, ok := w.itemExit(itemID)
		if !ok || (exit.Direction != direction && !item.matchesPhrase(itemID, direction)) {
			continue
		}
		if params, locked := item.tagParams("tag.lockable"); locked {
			keyID, _ := params["key_item_id"].(string)
			if keyID != "" && !w.playerHasItem(playerID, ItemID(keyID)) {
				return player.RoomID, false, "locked"
			}
		}
		player.RoomID = exit.TargetRoomID
		w.players[playerID] = player
		return exit.TargetRoomID, true, ""
	}
	return player.RoomID, false, ""
}

func (w *World) playerHasItem(playerID PlayerID, itemID ItemID) bool {
	for _, id := range w.itemsInInventory(playerID) {
		if id == itemID {
			return true
		}
	}
	return false
}
