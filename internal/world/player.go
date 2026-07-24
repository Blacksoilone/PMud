package world

func (w *World) EnterWorld(playerID PlayerID) (EntityID, bool) {
	if w.store.Player(playerID) != nil {
		return "", false
	}
	roomID := w.startRoom
	ent := &Entity{
		ID:     playerID,
		Name:   string(playerID),
		Player: &PlayerData{MaxWeight: 20, MaxVolume: 10},
	}
	w.store.Add(ent)
	w.store.PlaceInRoom(playerID, roomID)
	return roomID, true
}

func (w *World) LeaveWorld(playerID PlayerID) {
	w.store.Remove(playerID)
	delete(w.containerContents, PlayerContainerID(playerID))
	delete(w.trackedQuests, playerID)
}

func (w *World) PlayerCurrentRoom(playerID PlayerID) EntityID {
	return w.store.IsInRoom(playerID)
}

func (w *World) PlayerRoom(playerID PlayerID) (EntityID, bool) {
	roomID := w.store.IsInRoom(playerID)
	if roomID == "" || w.store.Player(playerID) == nil {
		return "", false
	}
	return roomID, true
}

func (w *World) PlayersInRoom(roomID EntityID) []PlayerID {
	ids := make([]PlayerID, 0)
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		if w.store.Player(eid) != nil {
			ids = append(ids, eid)
		}
	}
	return ids
}

func (w *World) PlayerCount() int {
	count := 0
	for _, eid := range w.store.Entities() {
		if w.store.Player(eid) != nil {
			count++
		}
	}
	return count
}

func (w *World) MovePlayer(playerID EntityID, direction string) (EntityID, bool) {
	roomID := w.store.IsInRoom(playerID)
	if roomID == "" {
		return "", false
	}
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		ed := w.store.Exit(eid)
		if ed == nil {
			continue
		}
		if ed.Direction == direction {
			w.store.PlaceInRoom(playerID, ed.TargetRoomID)
			return ed.TargetRoomID, true
		}
		if ent := w.store.Get(eid); ent != nil && ent.matchesPhrase(eid, direction) {
			w.store.PlaceInRoom(playerID, ed.TargetRoomID)
			return ed.TargetRoomID, true
		}
	}
	return roomID, false
}

func (w *World) PlayerHasItem(playerID EntityID, itemID EntityID) bool {
	for _, id := range w.containerContents[PlayerContainerID(playerID)] {
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
