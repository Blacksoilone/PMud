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
	Dark           bool
}

type ItemObservation struct {
	Item           ItemID
	NameKey        string
	DescriptionKey string
	Name           string
	Description    string
	Tags           []string // 可见 tag 描述文本
	PartTags       map[string][]string // part名 → 可见 tag 描述文本
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
	Dark           bool
}

type ItemPart struct {
	Tags []TagInstance
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

type ContainerItemLocation struct {
	ContainerID string
}

func (ContainerItemLocation) itemLocation() {}

// PlayerContainerID 返回玩家背包的容器 ID
func PlayerContainerID(pid PlayerID) string { return "player:" + string(pid) }

// ItemContainerID 返回物品容器（箱子/收纳袋）的容器 ID
func ItemContainerID(iid ItemID) string { return "item:" + string(iid) }

type PlayerEntity struct {
	ID        PlayerID
	RoomID    RoomID
	MaxWeight int
	MaxVolume int
}

type World struct {
	startRoom              RoomID
	rooms                  map[RoomID]Room
	items                  map[ItemID]Item
	itemLocations          map[ItemID]ItemLocation
	progressionDefinitions progression.Definitions
	players                map[PlayerID]PlayerEntity
	tagDefinitions         map[TagID]TagDefinition
	contentVerbs           map[string]VerbEntry
	trackedQuests          map[PlayerID]string
	containerOpen          map[ItemID]bool
	litItems               map[ItemID]bool
}
