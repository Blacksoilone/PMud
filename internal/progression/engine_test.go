package progression

import "testing"

func TestEngine_advancesTutorialQuestStages(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	// When — get lantern
	statuses := engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})

	// Then
	if len(statuses) != 1 {
		t.Fatalf("expected 1 advanced quest, got %d", len(statuses))
	}
	if statuses[0].StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("stage after get = %q", statuses[0].StageID)
	}
	if statuses[0].State != QuestStateActive {
		t.Fatalf("state after get = %q", statuses[0].State)
	}

	// When — move to yard
	statuses = engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})

	// Then
	if len(statuses) != 1 {
		t.Fatalf("expected 1 advanced quest, got %d", len(statuses))
	}
	if statuses[0].StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("stage after move = %q", statuses[0].StageID)
	}
	if statuses[0].State != QuestStateActive {
		t.Fatalf("state after move = %q", statuses[0].State)
	}

	// When — examine sword
	statuses = engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})

	// Then
	if len(statuses) != 1 {
		t.Fatalf("expected 1 advanced quest, got %d", len(statuses))
	}
	if statuses[0].StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("final stage id = %q", statuses[0].StageID)
	}
	if statuses[0].State != QuestStateRewardPending {
		t.Fatalf("final state = %q", statuses[0].State)
	}
}

func TestEngine_statusReportsCurrentStage(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())

	status, ok := engine.Status("player.local", "quest.tutorial.first_steps")

	if !ok {
		t.Fatal("expected tutorial quest status")
	}
	if status.QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("quest id = %q", status.QuestID)
	}
	if status.QuestName != "教程任务" {
		t.Fatalf("quest name = %q", status.QuestName)
	}
	if status.StageID != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage id = %q", status.StageID)
	}
	if status.StageText != "拿起旧油灯。" {
		t.Fatalf("stage text = %q", status.StageText)
	}
	if len(status.Conditions) != 1 || status.Conditions[0] != "获取旧油灯" {
		t.Fatalf("conditions = %#v", status.Conditions)
	}
	if status.State != QuestStateActive {
		t.Fatalf("state = %q", status.State)
	}
}

func TestEngine_ignoresUnmatchedTriggers(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	statuses := engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})

	if len(statuses) != 0 {
		t.Fatal("expected unmatched trigger not to advance any quest")
	}
}

func TestEngine_resolvesPendingRewardsToCompleted(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"
	engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})
	engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})
	statuses := engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})
	if len(statuses) != 1 || statuses[0].State != QuestStateRewardPending {
		t.Fatalf("state before reward resolution = %q", statuses[0].State)
	}

	status, resolved := engine.ResolveRewards(playerID, "quest.tutorial.first_steps")

	if !resolved {
		t.Fatal("expected pending rewards to resolve")
	}
	if status.State != QuestStateCompleted {
		t.Fatalf("state after reward resolution = %q", status.State)
	}
	if status.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("stage after reward resolution = %q", status.StageID)
	}
}

func TestEngine_doesNotResolveRewardsBeforeRewardPending(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	status, resolved := engine.ResolveRewards(playerID, "quest.tutorial.first_steps")

	if resolved {
		t.Fatal("expected active quest rewards not to resolve")
	}
	if status.State != QuestStateActive {
		t.Fatalf("state = %q", status.State)
	}
}

func TestEngine_resolvesTheRequestedQuestOnly(t *testing.T) {
	engine := NewEngine(Definitions{
		Quests: map[string]QuestDefinition{
			"quest.a": {ID: "quest.a", Name: "甲", StageIDs: []string{"stage.a"}},
			"quest.b": {ID: "quest.b", Name: "乙", StageIDs: []string{"stage.b"}},
		},
		Stages: map[string]StageDefinition{
			"stage.a": {ID: "stage.a", Text: "甲", Conditions: []ConditionDefinition{{Kind: string(TriggerGotItem), ItemID: "item.key", Text: "甲"}}},
			"stage.b": {ID: "stage.b", Text: "乙", Conditions: []ConditionDefinition{{Kind: string(TriggerGotItem), ItemID: "item.key", Text: "乙"}}},
		},
	})

	statuses := engine.Apply("player.local", Trigger{Kind: TriggerGotItem, ItemID: "item.key"})
	if len(statuses) != 2 {
		t.Fatalf("advanced statuses = %d, want 2", len(statuses))
	}

	status, resolved := engine.ResolveRewards("player.local", "quest.b")
	if !resolved || status.QuestID != "quest.b" || status.State != QuestStateCompleted {
		t.Fatalf("resolved = %#v, %v; want completed quest.b", status, resolved)
	}
	other, ok := engine.Status("player.local", "quest.a")
	if !ok || other.State != QuestStateRewardPending {
		t.Fatalf("quest.a = %#v, %v; want reward_pending", other, ok)
	}
}

func TestEngine_customConditionChecker_registersExternally(t *testing.T) {
	defs := Definitions{
		Quests: map[string]QuestDefinition{
			"quest.custom": {
				ID: "quest.custom", Name: "自定义",
				StageIDs: []string{"stage.kill"},
			},
		},
		Stages: map[string]StageDefinition{
			"stage.kill": {
				ID: "stage.kill", Text: "击杀 5 只老鼠。",
				Conditions: []ConditionDefinition{
					{Kind: "killed_monster", ItemID: "monster.rat", Text: "击杀老鼠 0/5"},
				},
			},
		},
	}
	engine := NewEngine(defs)

	engine.RegisterConditionChecker("killed_monster", func(c ConditionDefinition, t Trigger) bool {
		return t.Kind == TriggerGotItem && c.ItemID == t.ItemID
	})

	statuses := engine.Apply("p1", Trigger{Kind: TriggerGotItem, ItemID: "monster.rat"})

	if len(statuses) != 1 {
		t.Fatalf("expected 1 advanced quest, got %d", len(statuses))
	}
	if statuses[0].State != QuestStateRewardPending {
		t.Fatalf("state after kill = %q", statuses[0].State)
	}
}

func TestEngine_AllStatuses_returnsAllQuests(t *testing.T) {
	engine := NewEngine(tutorialDefinitions())

	statuses := engine.AllStatuses("player.local")

	if len(statuses) != 1 {
		t.Fatalf("expected 1 quest status, got %d", len(statuses))
	}
	if statuses[0].QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("quest id = %q", statuses[0].QuestID)
	}
}

func TestEngine_manualQuestStartsUnlockedAndActivatesExplicitly(t *testing.T) {
	defs := tutorialDefinitions()
	quest := defs.Quests["quest.tutorial.first_steps"]
	quest.Activation = ActivationManualAccept
	defs.Quests[quest.ID] = quest
	engine := NewEngine(defs)

	status, ok := engine.Status("player.local", quest.ID)
	if !ok || status.State != QuestStateUnlocked {
		t.Fatalf("initial status = %#v, %v; want unlocked", status, ok)
	}
	status, activated := engine.ActivateQuest("player.local", quest.ID)
	if !activated || status.State != QuestStateActive {
		t.Fatalf("activated status = %#v, %v; want active", status, activated)
	}
}

func TestEngine_autoOnEventQuestStartsHiddenAndActivatesOnMatchingEvent(t *testing.T) {
	defs := tutorialDefinitions()
	quest := defs.Quests["quest.tutorial.first_steps"]
	quest.Activation = ActivationAutoOnEvent
	quest.ActivationConditions = []ConditionDefinition{{Kind: string(TriggerMovedRoom), RoomID: "room.secret"}}
	defs.Quests[quest.ID] = quest
	engine := NewEngine(defs)

	status, _ := engine.Status("player.local", quest.ID)
	if status.State != QuestStateHidden {
		t.Fatalf("initial state = %q, want hidden", status.State)
	}
	if got := engine.Apply("player.local", Trigger{Kind: TriggerMovedRoom, RoomID: "room.other"}); len(got) != 0 {
		t.Fatalf("unmatched activation advanced = %#v", got)
	}
	got := engine.Apply("player.local", Trigger{Kind: TriggerMovedRoom, RoomID: "room.secret"})
	if len(got) != 1 || got[0].State != QuestStateActive {
		t.Fatalf("matching activation = %#v, want active", got)
	}
}

func TestEngine_repeatableQuestWaitsForRefreshAfterRewards(t *testing.T) {
	defs := tutorialDefinitions()
	quest := defs.Quests["quest.tutorial.first_steps"]
	quest.Repeatable = true
	defs.Quests[quest.ID] = quest
	engine := NewEngine(defs)
	playerID := "player.local"
	engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})
	engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})
	engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})

	status, resolved := engine.ResolveRewards(playerID, quest.ID)
	if !resolved || status.State != QuestStateWaitingRefresh {
		t.Fatalf("resolved repeatable quest = %#v, %v; want waiting_refresh", status, resolved)
	}
	status, refreshed := engine.RefreshQuest(playerID, quest.ID)
	if !refreshed || status.State != QuestStateActive || status.StageID != quest.StageIDs[0] {
		t.Fatalf("refreshed repeatable quest = %#v, %v; want first active stage", status, refreshed)
	}
}

func TestEngine_refreshQuestRestoresActivationState(t *testing.T) {
	for _, test := range []struct {
		name       string
		activation ActivationMode
		want       QuestState
	}{
		{name: "always active", activation: ActivationAlwaysActive, want: QuestStateActive},
		{name: "manual accept", activation: ActivationManualAccept, want: QuestStateUnlocked},
		{name: "auto on event", activation: ActivationAutoOnEvent, want: QuestStateHidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			defs := tutorialDefinitions()
			quest := defs.Quests["quest.tutorial.first_steps"]
			quest.Activation = test.activation
			quest.Repeatable = true
			defs.Quests[quest.ID] = quest
			engine := NewEngine(defs)
			runtime := engine.playerRuntime("player.local")
			state := runtime[quest.ID]
			state.state = QuestStateWaitingRefresh
			state.currentStageID = quest.StageIDs[len(quest.StageIDs)-1]
			runtime[quest.ID] = state

			status, refreshed := engine.RefreshQuest("player.local", quest.ID)
			if !refreshed || status.State != test.want || status.StageID != quest.StageIDs[0] {
				t.Fatalf("RefreshQuest() = %#v, %v; want state %q at first stage", status, refreshed, test.want)
			}
		})
	}
}

func tutorialDefinitions() Definitions {
	return Definitions{
		Quests: map[string]QuestDefinition{
			"quest.tutorial.first_steps": {
				ID:   "quest.tutorial.first_steps",
				Name: "教程任务",
				StageIDs: []string{
					"quest.tutorial.first_steps.stage.get_lantern",
					"quest.tutorial.first_steps.stage.enter_yard",
					"quest.tutorial.first_steps.stage.examine_sword",
				},
			},
		},
		Stages: map[string]StageDefinition{
			"quest.tutorial.first_steps.stage.get_lantern": {
				ID:     "quest.tutorial.first_steps.stage.get_lantern",
				Text:   "拿起旧油灯。",
				NextID: "quest.tutorial.first_steps.stage.enter_yard",
				Conditions: []ConditionDefinition{
					{Kind: string(TriggerGotItem), ItemID: "item.tutorial.old_lantern", Text: "获取旧油灯"},
				},
			},
			"quest.tutorial.first_steps.stage.enter_yard": {
				ID:     "quest.tutorial.first_steps.stage.enter_yard",
				Text:   "前往练习场。",
				NextID: "quest.tutorial.first_steps.stage.examine_sword",
				Conditions: []ConditionDefinition{
					{Kind: string(TriggerMovedRoom), RoomID: "room.tutorial.yard", Text: "到达练习场"},
				},
			},
			"quest.tutorial.first_steps.stage.examine_sword": {
				ID:   "quest.tutorial.first_steps.stage.examine_sword",
				Text: "查看练习木剑。",
				Conditions: []ConditionDefinition{
					{Kind: string(TriggerExaminedItem), ItemID: "item.tutorial.practice_sword", Text: "查看练习木剑"},
				},
			},
		},
	}
}
