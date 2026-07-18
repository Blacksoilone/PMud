package world

import "PMud/internal/presentation"

func newRoomObservationEvent(obs RoomObservation) presentation.Event {
	return presentation.RoomObservationEvent{
		Room:           string(obs.Room),
		NameKey:        obs.NameKey,
		DescriptionKey: obs.DescriptionKey,
		Name:           obs.Name,
		Description:    obs.Description,
		Exits:          obs.Exits,
		Neighbors:      roomNeighborStrings(obs.Neighbors),
		Items:          itemIDsToStrings(obs.ItemIDs),
	}
}

func newItemObservationEvent(obs ItemObservation) presentation.Event {
	return presentation.ItemObservationEvent{
		Item:           string(obs.Item),
		NameKey:        obs.NameKey,
		DescriptionKey: obs.DescriptionKey,
		Name:           obs.Name,
		Description:    obs.Description,
	}
}

func ambiguousItemsEvent(w *World, itemIDs []ItemID) []presentation.Event {
	return []presentation.Event{
		presentation.SystemMessageEvent{
			Message: "名字不明确: " + joinItemNames(w, itemIDs),
		},
	}
}

func roomNeighborStrings(neighbors map[string]RoomID) map[string]string {
	result := make(map[string]string, len(neighbors))
	for direction, roomID := range neighbors {
		result[direction] = string(roomID)
	}
	return result
}

func itemIDsToStrings(itemIDs []ItemID) []string {
	items := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, string(itemID))
	}
	return items
}

func joinItemNames(w *World, itemIDs []ItemID) string {
	names := w.ItemNames(itemIDs)
	result := ""
	for i, name := range names {
		if i > 0 {
			result += ", "
		}
		result += name
	}
	return result
}
