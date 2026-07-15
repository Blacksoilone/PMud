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
	DisplayNameKey TextKey
	InnerNameKey   TextKey
	DescriptionKey TextKey
	Aliases        []TextKey
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

type ServerItem struct {
	DisplayNameKey TextKey
	InnerNameKey   TextKey
	DescriptionKey TextKey
	Aliases        []TextKey
}

type ClientCatalog struct {
	RoomNames        map[RoomID]TextKey
	RoomDescriptions map[RoomID]TextKey
	ItemDisplayNames map[ItemID]TextKey
	ItemInnerNames   map[ItemID]TextKey
	ItemDescriptions map[ItemID]TextKey
	ItemAliases      map[ItemID][]TextKey
	Text             map[TextKey]string
}
