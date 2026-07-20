package content

type (
	RoomID             string
	ItemID             string
	TextKey            string
	Direction          string
	QuestID            string
	QuestStageID       string
	QuestConditionKind string
	TagID              string
	VerbID             string
)

const (
	TagExit      TagID = "exit"
	TagCarryable TagID = "carryable"
	TagLightable TagID = "lightable"
	TagContainer TagID = "container"
	TagLockable  TagID = "lockable"
)

const (
	QuestConditionGotItem      QuestConditionKind = "got_item"
	QuestConditionMovedRoom    QuestConditionKind = "moved_room"
	QuestConditionExaminedItem QuestConditionKind = "examined_item"
)

type VerbSource struct {
	ID          VerbID
	MessageKey  TextKey
}

type ContentSource struct {
	StartRoomID RoomID
	Rooms       []RoomSource
	Items       []ItemSource
	Verbs       []VerbSource
	Quests      []QuestSource
	QuestStages []QuestStageSource
	Text        map[TextKey]string
}

type RoomSource struct {
	ID             RoomID
	NameKey        TextKey
	DescriptionKey TextKey
	Dark           bool
}

type PartID string

type PartSource struct {
	SourceTags []SourceTag
}

type ItemSource struct {
	ID             ItemID
	DisplayNameKey TextKey
	InnerNameKey   TextKey
	DescriptionKey TextKey
	Aliases        []TextKey
	InitialRoom    RoomID
	Tags           []SourceTag
	Parts          map[PartID]PartSource
	Weight         int
	Volume         int
}

type SourceTag struct {
	ID     TagID
	Params map[string]string
}

type ServerPart struct {
	Tags []ServerTag
}

type RawTag struct {
	ID     TagID
	Params map[string]string
}

// ServerTag 使用鉴别联合体：命中的已知 tag 填入对应字段，
// 未知 tag 通过 GenericID/GenericParams 透传。
type ServerTag struct {
	Exit      *ExitTag
	Carryable bool
	Lightable bool
	Container *ContainerTag
	Lockable  *LockableTag
	// 通用透传：当 tag ID 不属于上述已知类型时存入此处
	GenericID     TagID
	GenericParams map[string]string
}

type QuestSource struct {
	ID       QuestID
	NameKey  TextKey
	StageIDs []QuestStageID
}

type QuestStageSource struct {
	ID               QuestStageID
	TextKey          TextKey
	FinishConditions []QuestConditionSource
	NextStageID      QuestStageID
}

type QuestConditionSource struct {
	Kind   QuestConditionKind
	ItemID ItemID
	RoomID RoomID
}

type CompiledContent struct {
	Server ServerSnapshot
	Client ClientCatalog
}

type ServerVerb struct {
	MessageKey TextKey
}

type ServerSnapshot struct {
	StartRoomID   RoomID
	Rooms         map[RoomID]ServerRoom
	Items         map[ItemID]ServerItem
	ItemLocations map[ItemID]RoomID
	Verbs         map[VerbID]ServerVerb
	Quests        map[QuestID]ServerQuest
	QuestStages   map[QuestStageID]ServerQuestStage
}

type ServerRoom struct {
	Dark bool
}

type ServerItem struct {
	DisplayNameKey TextKey
	InnerNameKey   TextKey
	DescriptionKey TextKey
	Aliases        []TextKey
	Tags           []ServerTag
	Parts          map[PartID]ServerPart
	Weight         int
	Volume         int
}

type ExitTag struct {
	Direction    Direction
	TargetRoomID RoomID
}

type ContainerTag struct {
	Capacity int
}

type LockableTag struct {
	KeyItemID ItemID
}

type ServerQuest struct {
	NameKey  TextKey
	StageIDs []QuestStageID
}

type ServerQuestStage struct {
	TextKey          TextKey
	FinishConditions []ServerQuestCondition
	NextStageID      QuestStageID
}

type ServerQuestCondition struct {
	Kind   QuestConditionKind
	ItemID ItemID
	RoomID RoomID
}

type ClientCatalog struct {
	RoomNames        map[RoomID]TextKey
	RoomDescriptions map[RoomID]TextKey
	ItemDisplayNames map[ItemID]TextKey
	ItemInnerNames   map[ItemID]TextKey
	ItemDescriptions map[ItemID]TextKey
	ItemAliases      map[ItemID][]TextKey
	VerbNames        map[VerbID]TextKey
	Text             map[TextKey]string
}
