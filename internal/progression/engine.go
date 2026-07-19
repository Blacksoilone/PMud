package progression

type Engine struct {
	definitions Definitions
	runtime     map[string]questRuntime
	checkers    map[string]ConditionChecker
}

func NewEngine(definitions Definitions) *Engine {
	e := &Engine{
		definitions: definitions,
		runtime:     make(map[string]questRuntime),
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

func (e *Engine) Apply(playerID string, trigger Trigger) (Status, bool) {
	runtime := e.runtimeFor(playerID)
	if runtime.state != QuestStateActive {
		return e.status(runtime), false
	}
	stage := e.definitions.Stages[runtime.currentStageID]
	if !stage.matches(trigger, e.checkers) {
		return e.status(runtime), false
	}
	if stage.NextID == "" {
		runtime.state = QuestStateRewardPending
	} else {
		runtime.currentStageID = stage.NextID
	}
	e.runtime[playerID] = runtime
	return e.status(runtime), true
}

func (e *Engine) Status(playerID string) (Status, bool) {
	if e.definitions.Quest.ID == "" || len(e.definitions.Quest.StageIDs) == 0 {
		return Status{}, false
	}
	return e.status(e.runtimeFor(playerID)), true
}

func (e *Engine) ResolveRewards(playerID string) (Status, bool) {
	runtime := e.runtimeFor(playerID)
	if runtime.state != QuestStateRewardPending {
		return e.status(runtime), false
	}
	runtime.state = QuestStateCompleted
	e.runtime[playerID] = runtime
	return e.status(runtime), true
}

func (e *Engine) runtimeFor(playerID string) questRuntime {
	if runtime, ok := e.runtime[playerID]; ok {
		return runtime
	}
	runtime := questRuntime{
		currentStageID: e.definitions.Quest.StageIDs[0],
		state:          QuestStateActive,
	}
	e.runtime[playerID] = runtime
	return runtime
}

func (e *Engine) status(runtime questRuntime) Status {
	stage := e.definitions.Stages[runtime.currentStageID]
	conditions := make([]string, 0, len(stage.Conditions))
	for _, condition := range stage.Conditions {
		conditions = append(conditions, condition.Text)
	}
	return Status{
		QuestID:    e.definitions.Quest.ID,
		QuestName:  e.definitions.Quest.Name,
		StageID:    runtime.currentStageID,
		StageText:  stage.Text,
		Conditions: conditions,
		State:      runtime.state,
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
