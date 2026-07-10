package world

func (w *World) StartRoom() RoomID {
	return w.startRoom
}

func (w *World) Look(roomID RoomID) (RoomObservation, bool) {
	room, ok := w.rooms[roomID]
	if !ok {
		return RoomObservation{}, false
	}

	exits := make([]string, 0, len(room.Exits))
	for direction := range room.Exits {
		exits = append(exits, direction)
	}
	items := w.itemNames(w.itemsInRoom(roomID))

	return RoomObservation{
		Name:        room.Name,
		Description: room.Description,
		Exits:       exits,
		Items:       items,
	}, true
}

func (w *World) Move(roomID RoomID, direction string) (RoomID, bool) {
	room, ok := w.rooms[roomID]
	if !ok {
		return roomID, false
	}

	nextRoomID, ok := room.Exits[direction]
	if !ok {
		return roomID, false
	}

	return nextRoomID, true
}
