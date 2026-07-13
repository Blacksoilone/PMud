package content

type RoomID string
type ItemID string
type TextKey string
type Direction string

type ContentSource struct {
	StartRoomID RoomID
	Rooms       []RoomSource
	Items       []ItemSource
	Text        map[TextKey]string
}

type RoomSource struct {
	ID             RoomID
	NameKey        TextKey
	DescriptionKey TextKey
	Exits          map[Direction]RoomID
}

type ItemSource struct {
	ID             ItemID
	NameKey        TextKey
	DescriptionKey TextKey
	InitialRoom    RoomID
}

type CompiledContent struct {
	Server ServerSnapshot
	Client ClientCatalog
}

type ServerSnapshot struct {
	StartRoomID   RoomID
	Rooms         map[RoomID]ServerRoom
	Items         map[ItemID]ServerItem
	ItemLocations map[ItemID]RoomID
}

type ServerRoom struct {
	Exits map[Direction]RoomID
}

type ServerItem struct{}

type ClientCatalog struct {
	RoomNames        map[RoomID]TextKey
	RoomDescriptions map[RoomID]TextKey
	ItemNames        map[ItemID]TextKey
	ItemDescriptions map[ItemID]TextKey
	Text             map[TextKey]string
}
