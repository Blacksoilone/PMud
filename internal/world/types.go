package world

import "PMud/internal/progression"

type RoomObservation struct {
	Room           RoomID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Exits          []string
	Neighbors      map[string]RoomID
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

type (
	RoomID   string
	ItemID   string
	PlayerID string
)

type Room struct {
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
}

type Item struct {
	NameKey        string
	InnerName      string
	DescriptionKey string
	Name           string
	Description    string
	Aliases        []string
	Tags           []TagInstance
}

// Exit 是一个纯值类型，从 tag.exit 的 TagInstance 中提取
type Exit struct {
	Direction    string
	TargetRoomID RoomID
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

type PlayerEntity struct {
	ID     PlayerID
	RoomID RoomID
}

type World struct {
	startRoom              RoomID
	rooms                  map[RoomID]Room
	items                  map[ItemID]Item
	itemLocations          map[ItemID]ItemLocation
	progressionDefinitions progression.Definitions
	players                map[PlayerID]PlayerEntity
	tagDefinitions         map[TagID]TagDefinition
}
