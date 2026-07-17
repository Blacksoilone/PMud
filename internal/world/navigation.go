package world

func (w *World) StartRoom() RoomID {
	return w.startRoom
}

func (w *World) Look(roomID RoomID) (RoomObservation, bool) {
	room, ok := w.rooms[roomID]
	if !ok {
		return RoomObservation{}, false
	}

	exits := make([]string, 0)
	neighbors := make(map[string]RoomID)
	for _, itemID := range w.exitItemIDs(roomID) {
		exit, ok := w.itemExit(itemID)
		if ok && exit.Direction != "" {
			exits = append(exits, exit.Direction)
			if isPlanarDirection(exit.Direction) {
				neighbors[exit.Direction] = exit.TargetRoomID
			}
		}
	}
	itemIDs := w.ordinaryItemsInRoom(roomID)
	items := w.itemNames(itemIDs)

	return RoomObservation{
		Room:           roomID,
		NameKey:        room.NameKey,
		DescriptionKey: room.DescriptionKey,
		Name:           room.Name,
		Description:    room.Description,
		Exits:          exits,
		Neighbors:      neighbors,
		ItemIDs:        itemIDs,
		Items:          items,
	}, true
}

func isPlanarDirection(direction string) bool {
	switch direction {
	case "north", "northeast", "east", "southeast", "south", "southwest", "west", "northwest":
		return true
	default:
		return false
	}
}

func (w *World) Move(roomID RoomID, direction string) (RoomID, bool) {
	if _, ok := w.rooms[roomID]; !ok {
		return roomID, false
	}
	for _, itemID := range w.exitItemIDs(roomID) {
		item := w.items[itemID]
		exit, ok := w.itemExit(itemID)
		if !ok || (exit.Direction != direction && !item.matchesPhrase(itemID, direction)) {
			continue
		}
		return exit.TargetRoomID, true
	}
	return roomID, false
}
