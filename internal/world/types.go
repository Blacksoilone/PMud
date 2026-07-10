package world

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
