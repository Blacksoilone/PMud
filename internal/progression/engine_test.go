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
					{Kind: TriggerGotItem, ItemID: "item.tutorial.old_lantern", Text: "获取旧油灯"},
				},
			},
			"quest.tutorial.first_steps.stage.enter_yard": {
				ID:     "quest.tutorial.first_steps.stage.enter_yard",
				Text:   "前往练习场。",
				NextID: "quest.tutorial.first_steps.stage.examine_sword",
				Conditions: []ConditionDefinition{
					{Kind: TriggerMovedRoom, RoomID: "room.tutorial.yard", Text: "到达练习场"},
				},
			},
			"quest.tutorial.first_steps.stage.examine_sword": {
				ID:   "quest.tutorial.first_steps.stage.examine_sword",
				Text: "查看练习木剑。",
				Conditions: []ConditionDefinition{
					{Kind: TriggerExaminedItem, ItemID: "item.tutorial.practice_sword", Text: "查看练习木剑"},
				},
			},
		},
	}
}
