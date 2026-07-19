package progression

import "sort"

type Engine struct {
	definitions Definitions
	// runtime[playerID][questID] = runtime
	runtime  map[string]map[string]questRuntime
	checkers map[string]ConditionChecker
}

func NewEngine(definitions Definitions) *Engine {
	e := &Engine{
		definitions: definitions,
		runtime:     make(map[string]map[string]questRuntime),
		checkers:    make(map[string]ConditionChecker),
	}
	e.registerBuiltinCheckers()
	return e
}

func (e *Engine) RegisterConditionChecker(kind string, checker ConditionChecker) {
	e.checkers[kind] = checker
}

func (e *Engine) registerBuiltinCheckers() {
	e.checkers[string(TriggerGotItem)] = func(c ConditionDefinition, t Trigger) bool {
		return t.Kind == TriggerGotItem && (c.ItemID == "" || c.ItemID == t.ItemID)
	}
	e.checkers[string(TriggerMovedRoom)] = func(c ConditionDefinition, t Trigger) bool {
		return t.Kind == TriggerMovedRoom && (c.RoomID == "" || c.RoomID == t.RoomID)
	}
	e.checkers[string(TriggerExaminedItem)] = func(c ConditionDefinition, t Trigger) bool {
		return t.Kind == TriggerExaminedItem && (c.ItemID == "" || c.ItemID == t.ItemID)
	}
}

// questIDs returns sorted quest IDs from definitions.
func (e *Engine) questIDs() []string {
	ids := make([]string, 0, len(e.definitions.Quests))
	for id := range e.definitions.Quests {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Apply checks all active quests for the given player and advances any
// whose current stage conditions are satisfied by the trigger.
// Returns statuses for every quest that advanced.
func (e *Engine) Apply(playerID string, trigger Trigger) []Status {
	playerRuntime := e.playerRuntime(playerID)
	var advanced []Status
	for _, questID := range e.questIDs() {
		rt := playerRuntime[questID]
		if rt.state != QuestStateActive {
			continue
		}
		stage := e.definitions.Stages[rt.currentStageID]
		if !stage.matches(trigger, e.checkers) {
			continue
		}
		if stage.NextID == "" {
			rt.state = QuestStateRewardPending
		} else {
			rt.currentStageID = stage.NextID
		}
		playerRuntime[questID] = rt
		advanced = append(advanced, e.status(questID, rt))
	}
	if len(advanced) > 0 {
		e.runtime[playerID] = playerRuntime
	}
	return advanced
}

// AllStatuses returns the full status list for every quest.
func (e *Engine) AllStatuses(playerID string) []Status {
	playerRuntime := e.playerRuntime(playerID)
	result := make([]Status, 0, len(e.definitions.Quests))
	for _, questID := range e.questIDs() {
		rt := playerRuntime[questID]
		result = append(result, e.status(questID, rt))
	}
	return result
}

// Status returns the status for a single quest.
func (e *Engine) Status(playerID string, questID string) (Status, bool) {
	playerRuntime := e.playerRuntime(playerID)
	rt, ok := playerRuntime[questID]
	if !ok {
		return Status{}, false
	}
	return e.status(questID, rt), true
}

func (e *Engine) ResolveRewards(playerID string) (Status, bool) {
	playerRuntime := e.playerRuntime(playerID)
	for _, questID := range e.questIDs() {
		rt := playerRuntime[questID]
		if rt.state != QuestStateRewardPending {
			continue
		}
		rt.state = QuestStateCompleted
		playerRuntime[questID] = rt
		e.runtime[playerID] = playerRuntime
		return e.status(questID, rt), true
	}
	ids := e.questIDs()
	if len(ids) > 0 {
		return e.status(ids[0], playerRuntime[ids[0]]), false
	}
	return Status{}, false
}

func (e *Engine) playerRuntime(playerID string) map[string]questRuntime {
	if rt, ok := e.runtime[playerID]; ok {
		return rt
	}
	rt := make(map[string]questRuntime, len(e.definitions.Quests))
	for _, questID := range e.questIDs() {
		quest := e.definitions.Quests[questID]
		if len(quest.StageIDs) == 0 {
			continue
		}
		rt[questID] = questRuntime{
			currentStageID: quest.StageIDs[0],
			state:          QuestStateActive,
		}
	}
	e.runtime[playerID] = rt
	return rt
}

func (e *Engine) status(questID string, rt questRuntime) Status {
	quest := e.definitions.Quests[questID]
	stage := e.definitions.Stages[rt.currentStageID]
	conditions := make([]string, 0, len(stage.Conditions))
	for _, condition := range stage.Conditions {
		conditions = append(conditions, condition.Text)
	}
	return Status{
		QuestID:    questID,
		QuestName:  quest.Name,
		StageID:    rt.currentStageID,
		StageText:  stage.Text,
		Conditions: conditions,
		State:      rt.state,
	}
}

func (s StageDefinition) matches(trigger Trigger, checkers map[string]ConditionChecker) bool {
	if len(s.Conditions) == 0 {
		return false
	}
	for _, condition := range s.Conditions {
		checker, ok := checkers[condition.Kind]
		if !ok {
			return false
		}
		if !checker(condition, trigger) {
			return false
		}
	}
	return true
}
