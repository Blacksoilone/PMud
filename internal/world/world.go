package world

import "PMud/internal/content"

type RoomObservation struct {
	Name        string
	Description string
	Exits       []string
	Items       []string
}

type RoomID string
type ItemID string
type PlayerID string

type Room struct {
	Name        string
	Description string
	Exits       map[string]RoomID
}

type Item struct {
	Name string
}

type ItemLocation interface {
	itemLocation()
}

type RoomItemLocation struct {
	RoomID RoomID
}

func (RoomItemLocation) itemLocation() {}

type InventoryItemLocation struct {
	PlayerID PlayerID
}

func (InventoryItemLocation) itemLocation() {}

type World struct {
	startRoom     RoomID
	rooms         map[RoomID]Room
	items         map[ItemID]Item
	itemLocations map[ItemID]ItemLocation
}

func New() *World {
	return &World{
		startRoom: "room.tutorial.start",
		rooms: map[RoomID]Room{
			"room.tutorial.start": {
				Name:        "练习场入口",
				Description: "这里是练习场的入口。北边传来木剑碰撞的声音。",
				Exits: map[string]RoomID{
					"north": "room.tutorial.yard",
				},
			},
			"room.tutorial.yard": {
				Name:        "练习场",
				Description: "几根木桩立在泥地上，地面满是被踩出的脚印。",
				Exits: map[string]RoomID{
					"south": "room.tutorial.start",
				},
			},
		},
		items: map[ItemID]Item{
			"item.tutorial.old_lantern": {
				Name: "旧油灯",
			},
			"item.tutorial.practice_sword": {
				Name: "练习木剑",
			},
		},
		itemLocations: map[ItemID]ItemLocation{
			"item.tutorial.old_lantern": RoomItemLocation{
				RoomID: "room.tutorial.start",
			},
			"item.tutorial.practice_sword": RoomItemLocation{
				RoomID: "room.tutorial.yard",
			},
		},
	}
}

func NewFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) *World {
	rooms := make(map[RoomID]Room, len(snapshot.Rooms))
	for roomID, room := range snapshot.Rooms {
		exits := make(map[string]RoomID, len(room.Exits))
		for direction, targetRoomID := range room.Exits {
			exits[string(direction)] = RoomID(targetRoomID)
		}

		nameKey := catalog.RoomNames[roomID]
		descriptionKey := catalog.RoomDescriptions[roomID]
		rooms[RoomID(roomID)] = Room{
			Name:        catalog.Text[nameKey],
			Description: catalog.Text[descriptionKey],
			Exits:       exits,
		}
	}

	items := make(map[ItemID]Item, len(snapshot.Items))
	for itemID := range snapshot.Items {
		nameKey := catalog.ItemNames[itemID]
		items[ItemID(itemID)] = Item{Name: catalog.Text[nameKey]}
	}

	itemLocations := make(map[ItemID]ItemLocation, len(snapshot.ItemLocations))
	for itemID, roomID := range snapshot.ItemLocations {
		itemLocations[ItemID(itemID)] = RoomItemLocation{RoomID: RoomID(roomID)}
	}

	return &World{
		startRoom:     RoomID(snapshot.StartRoomID),
		rooms:         rooms,
		items:         items,
		itemLocations: itemLocations,
	}
}

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

func (w *World) GetItem(roomID RoomID, name string, playerID PlayerID) (ItemID, bool) {
	for _, itemID := range w.itemsInRoom(roomID) {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		if item.Name != name {
			continue
		}

		w.itemLocations[itemID] = InventoryItemLocation{PlayerID: playerID}
		return itemID, true
	}

	return "", false
}

func (w *World) DropItem(roomID RoomID, itemID ItemID) {
	w.itemLocations[itemID] = RoomItemLocation{RoomID: roomID}
}

func (w *World) DropItemByName(roomID RoomID, name string, playerID PlayerID) bool {
	for _, itemID := range w.itemsInInventory(playerID) {
		item, ok := w.items[itemID]
		if !ok {
			continue
		}
		if item.Name != name {
			continue
		}

		w.DropItem(roomID, itemID)
		return true
	}

	return false
}

func (w *World) ItemName(itemID ItemID) (string, bool) {
	item, ok := w.items[itemID]
	if !ok {
		return "", false
	}
	return item.Name, true
}

func (w *World) ItemNames(itemIDs []ItemID) []string {
	return w.itemNames(itemIDs)
}

func (w *World) Inventory(playerID PlayerID) []string {
	return w.itemNames(w.itemsInInventory(playerID))
}

func (w *World) itemNames(itemIDs []ItemID) []string {
	names := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		name, ok := w.ItemName(itemID)
		if !ok {
			continue
		}
		names = append(names, name)
	}
	return names
}

func (w *World) itemsInRoom(roomID RoomID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for itemID, location := range w.itemLocations {
		roomLocation, ok := location.(RoomItemLocation)
		if !ok {
			continue
		}
		if roomLocation.RoomID == roomID {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}

func (w *World) itemsInInventory(playerID PlayerID) []ItemID {
	itemIDs := make([]ItemID, 0)
	for itemID, location := range w.itemLocations {
		inventoryLocation, ok := location.(InventoryItemLocation)
		if !ok {
			continue
		}
		if inventoryLocation.PlayerID == playerID {
			itemIDs = append(itemIDs, itemID)
		}
	}
	return itemIDs
}
