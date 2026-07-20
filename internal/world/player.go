package world

func (w *World) EnterWorld(playerID PlayerID) (RoomID, bool) {
	if _, exists := w.players[playerID]; exists {
		return "", false
	}
	w.players[playerID] = PlayerEntity{
		ID:        playerID,
		RoomID:    w.startRoom,
		MaxWeight: 20,
		MaxVolume: 10,
	}
	return w.startRoom, true
}

func (w *World) LeaveWorld(playerID PlayerID) {
	delete(w.players, playerID)
	delete(w.trackedQuests, playerID)
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

func (w *World) MovePlayer(playerID PlayerID, direction string) (RoomID, bool) {
	player, ok := w.players[playerID]
	if !ok {
		return "", false
	}
	for _, itemID := range w.exitItemIDs(player.RoomID) {
		exit, ok := w.itemExit(itemID)
		if !ok || exit.Direction != direction {
			continue
		}
		player.RoomID = exit.TargetRoomID
		w.players[playerID] = player
		return exit.TargetRoomID, true
	}
	return player.RoomID, false
}

func (w *World) PlayerHasItem(playerID PlayerID, itemID ItemID) bool {
	for _, id := range w.itemsInContainer(PlayerContainerID(playerID)) {
		if id == itemID {
			return true
		}
	}
	return false
}

func (w *World) TrackedQuest(playerID PlayerID) string {
	if id, ok := w.trackedQuests[playerID]; ok && id != "" {
		return id
	}
	return ""
}

func (w *World) SetTrackedQuest(playerID PlayerID, questID string) {
	w.trackedQuests[playerID] = questID
}
