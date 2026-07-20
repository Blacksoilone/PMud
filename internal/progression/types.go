package progression

type TriggerKind string

const (
	TriggerGotItem      TriggerKind = "got_item"
	TriggerMovedRoom    TriggerKind = "moved_room"
	TriggerExaminedItem TriggerKind = "examined_item"
)

type QuestState string

const (
	QuestStateHidden         QuestState = "hidden"
	QuestStateUnlocked       QuestState = "unlocked"
	QuestStateActive         QuestState = "active"
	QuestStateRewardPending  QuestState = "reward_pending"
	QuestStateCompleted      QuestState = "completed"
	QuestStateWaitingRefresh QuestState = "waiting_refresh"
	QuestStateRetryWait      QuestState = "retry_wait"
)

type ActivationMode string

const (
	ActivationManualAccept ActivationMode = "manual_accept"
	ActivationAutoOnEvent  ActivationMode = "auto_on_event"
	ActivationAlwaysActive ActivationMode = "always_active"
)

type Trigger struct {
	Kind   TriggerKind
	ItemID string
	RoomID string
}

// Definitions contains all quest definitions and their stages.
type Definitions struct {
	Quests map[string]QuestDefinition // questID → definition
	Stages map[string]StageDefinition
}

type QuestDefinition struct {
	ID                   string
	Name                 string
	StageIDs             []string
	Activation           ActivationMode
	ActivationConditions []ConditionDefinition
	Repeatable           bool
}

type StageDefinition struct {
	ID         string
	Text       string
	Conditions []ConditionDefinition
	NextID     string
}

type ConditionChecker func(condition ConditionDefinition, trigger Trigger) bool

type ConditionDefinition struct {
	Kind   string
	ItemID string
	RoomID string
	Text   string
}

// Status represents a player's progress on a single quest.
type Status struct {
	QuestID    string
	QuestName  string
	StageID    string
	StageText  string
	Conditions []string
	State      QuestState
}

type questRuntime struct {
	currentStageID string
	state          QuestState
}
