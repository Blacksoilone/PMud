package world

type RoomObservation struct {
	Room           RoomID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Exits          []string
	ItemIDs        []ItemID
	Items          []string
}

type ItemObservation struct {
	Item           ItemID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
}

type RoomID string
type ItemID string
type PlayerID string

type Room struct {
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Exits          map[string]RoomID
}

type Item struct {
	NameKey        string
	InnerName      string
	DescriptionKey string
	Name           string
	Description    string
	Aliases        []string
}

type ItemResolution struct {
	ItemID           ItemID
	AmbiguousItemIDs []ItemID
	Found            bool
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
