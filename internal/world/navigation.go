package world

func (w *World) StartRoom() EntityID {
	return w.startRoom
}

func (w *World) Look(roomID EntityID) (RoomObservation, bool) {
	rd := w.store.Room(roomID)
	if rd == nil {
		return RoomObservation{}, false
	}
	ent := w.store.Get(roomID)

	exits := make([]string, 0)
	neighbors := make(map[string]EntityID)
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		ed := w.store.Exit(eid)
		if ed == nil || ed.Direction == "" {
			continue
		}
		exits = append(exits, ed.Direction)
		if isPlanarDirection(ed.Direction) {
			neighbors[ed.Direction] = ed.TargetRoomID
		}
	}
	itemIDs := w.ordinaryItemsInRoom(roomID)
	items := w.itemNames(itemIDs)

	return RoomObservation{
		Room:           roomID,
		NameKey:        ent.NameKey,
		DescriptionKey: ent.DescriptionKey,
		Name:           ent.Name,
		Description:    ent.Description,
		Exits:          exits,
		Neighbors:      neighbors,
		ItemIDs:        itemIDs,
		Items:          items,
		Dark:           rd.Dark,
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

func (w *World) Move(roomID EntityID, direction string) (EntityID, bool) {
	if w.store.Room(roomID) == nil {
		return roomID, false
	}
	for _, eid := range w.store.EntitiesInRoom(roomID) {
		ed := w.store.Exit(eid)
		if ed == nil {
			continue
		}
		if ed.Direction == direction {
			return ed.TargetRoomID, true
		}
		ent := w.store.Get(eid)
		if ent != nil && ent.matchesPhrase(eid, direction) {
			return ed.TargetRoomID, true
		}
	}
	return roomID, false
}
