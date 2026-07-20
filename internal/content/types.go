package content

type (
	RoomID             string
	ItemID             string
	TextKey            string
	Direction          string
	QuestID            string
	QuestStageID       string
	QuestConditionKind string
	QuestActivation    string
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

const (
	QuestActivationManualAccept QuestActivation = "manual_accept"
	QuestActivationAutoOnEvent  QuestActivation = "auto_on_event"
	QuestActivationAlwaysActive QuestActivation = "always_active"
)

type VerbSource struct {
	ID         VerbID
	MessageKey TextKey
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

type ServerTag struct {
	Exit      *ExitTag
	Carryable bool
	Lightable bool
	Container *ContainerTag
	Lockable  *LockableTag
}

type QuestSource struct {
	ID                   QuestID
	NameKey              TextKey
	StageIDs             []QuestStageID
	Activation           QuestActivation
	ActivationConditions []QuestConditionSource
	Repeatable           bool
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
	NameKey              TextKey
	StageIDs             []QuestStageID
	Activation           QuestActivation
	ActivationConditions []ServerQuestCondition
	Repeatable           bool
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
