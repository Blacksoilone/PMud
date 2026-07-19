package progression

import "testing"

func TestEngine_advancesTutorialQuestStages(t *testing.T) {
	// Given
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	// When
	status, advanced := engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})

	// Then
	if !advanced {
		t.Fatal("expected first trigger to advance quest")
	}
	if status.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("stage after get = %q", status.StageID)
	}
	if status.State != QuestStateActive {
		t.Fatalf("state after get = %q", status.State)
	}

	// When
	status, advanced = engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})

	// Then
	if !advanced {
		t.Fatal("expected move trigger to advance quest")
	}
	if status.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("stage after move = %q", status.StageID)
	}
	if status.State != QuestStateActive {
		t.Fatalf("state after move = %q", status.State)
	}

	// When
	status, advanced = engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})

	// Then
	if !advanced {
		t.Fatal("expected examine trigger to advance quest")
	}
	if status.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("final stage id = %q", status.StageID)
	}
	if status.State != QuestStateRewardPending {
		t.Fatalf("final state = %q", status.State)
	}
}

func TestEngine_statusReportsCurrentStage(t *testing.T) {
	// Given
	engine := NewEngine(tutorialDefinitions())

	// When
	status, ok := engine.Status("player.local")

	// Then
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
	// Given
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	// When
	status, advanced := engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})

	// Then
	if advanced {
		t.Fatal("expected unmatched trigger not to advance quest")
	}
	if status.StageID != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage id = %q", status.StageID)
	}
}

func TestEngine_resolvesPendingRewardsToCompleted(t *testing.T) {
	// Given
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"
	engine.Apply(playerID, Trigger{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern"})
	engine.Apply(playerID, Trigger{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard"})
	status, _ := engine.Apply(playerID, Trigger{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword"})
	if status.State != QuestStateRewardPending {
		t.Fatalf("state before reward resolution = %q", status.State)
	}

	// When
	status, resolved := engine.ResolveRewards(playerID)

	// Then
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
	// Given
	engine := NewEngine(tutorialDefinitions())
	playerID := "player.local"

	// When
	status, resolved := engine.ResolveRewards(playerID)

	// Then
	if resolved {
		t.Fatal("expected active quest rewards not to resolve")
	}
	if status.State != QuestStateActive {
		t.Fatalf("state = %q", status.State)
	}
}

func TestEngine_customConditionChecker_registersExternally(t *testing.T) {
	// Given — a quest with a custom condition kind
	defs := Definitions{
		Quest: QuestDefinition{
			ID: "quest.custom", Name: "自定义",
			StageIDs: []string{"stage.kill"},
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

	// Register a custom checker — 5 rat kills required
	engine.RegisterConditionChecker("killed_monster", func(c ConditionDefinition, t Trigger) bool {
		return t.Kind == TriggerGotItem && c.ItemID == t.ItemID
	})

	// When — first kill
	status, advanced := engine.Apply("p1", Trigger{Kind: TriggerGotItem, ItemID: "monster.rat"})

	// Then
	if !advanced {
		t.Fatal("expected custom condition to match on first kill")
	}
	if status.State != QuestStateRewardPending {
		t.Fatalf("state after kill = %q", status.State)
	}
}

func tutorialDefinitions() Definitions {
	return Definitions{
		Quest: QuestDefinition{
			ID:   "quest.tutorial.first_steps",
			Name: "教程任务",
			StageIDs: []string{
				"quest.tutorial.first_steps.stage.get_lantern",
				"quest.tutorial.first_steps.stage.enter_yard",
				"quest.tutorial.first_steps.stage.examine_sword",
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
