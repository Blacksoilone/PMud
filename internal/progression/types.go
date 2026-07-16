package progression

type TriggerKind string

const (
	TriggerGotItem      TriggerKind = "got_item"
	TriggerMovedRoom    TriggerKind = "moved_room"
	TriggerExaminedItem TriggerKind = "examined_item"
)

type QuestState string

const (
	QuestStateActive        QuestState = "active"
	QuestStateRewardPending QuestState = "reward_pending"
	QuestStateCompleted     QuestState = "completed"
)

type Trigger struct {
	Kind   TriggerKind
	ItemID string
	RoomID string
}

type Definitions struct {
	Quest  QuestDefinition
	Stages map[string]StageDefinition
}

type QuestDefinition struct {
	ID       string
	Name     string
	StageIDs []string
}

type StageDefinition struct {
	ID         string
	Text       string
	Conditions []ConditionDefinition
	NextID     string
}

type ConditionDefinition struct {
	Kind   TriggerKind
	ItemID string
	RoomID string
	Text   string
}

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
