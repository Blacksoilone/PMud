package world

import "PMud/internal/progression"

type RoomObservation struct {
	Room           EntityID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Exits          []string
	Neighbors      map[string]EntityID
	ItemIDs        []EntityID
	Items          []string
	Dark           bool
}

type ItemObservation struct {
	Item           EntityID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Tags           []string
	PartTags       map[string][]string
}

type (
	RoomID   = EntityID
	ItemID   = EntityID
	PlayerID = EntityID
)

func PlayerContainerID(pid PlayerID) string { return "player:" + string(pid) }

func ItemContainerID(iid ItemID) string { return "item:" + string(iid) }

type ItemResolution struct {
	ItemID           EntityID
	AmbiguousItemIDs []EntityID
	Found            bool
}

type Exit struct {
	Direction    string
	TargetRoomID RoomID
}

type PlayerEntity struct {
	ID        PlayerID
	RoomID    RoomID
	MaxWeight int
	MaxVolume int
	Tags      []TagInstance
}

type World struct {
	store                  *EntityStore
	startRoom              EntityID
	progressionDefinitions progression.Definitions
	tagDefinitions         map[TagID]TagDefinition
	contentVerbs           map[string]VerbEntry
	containerContents      map[string][]EntityID
	containerOpen          map[EntityID]bool
	litItems               map[EntityID]bool
	trackedQuests          map[PlayerID]string
}

type Item struct {
	NameKey        string
	InnerName      string
	DescriptionKey string
	Name           string
	Description    string
	Aliases        []string
	Tags           []TagInstance
	Parts          map[string]ItemPart
	Weight         int
	Volume         int
}

type ItemPart struct {
	Tags []TagInstance
}
