package world

import (
	"slices"

	"PMud/internal/content"
	"PMud/internal/progression"
)

func New() *World {
	w := &World{
		startRoom: "room.tutorial.start",
		rooms: map[RoomID]Room{
			"room.tutorial.start": {
				NameKey:        "room.tutorial.start.name",
				DescriptionKey: "room.tutorial.start.description",
				Name:           "练习场入口",
				Description:    "这里是练习场的入口。北边传来木剑碰撞的声音。",
			},
			"room.tutorial.yard": {
				NameKey:        "room.tutorial.yard.name",
				DescriptionKey: "room.tutorial.yard.description",
				Name:           "练习场",
				Description:    "几根木桩立在泥地上，地面满是被踩出的脚印。",
			},
			"room.tutorial.shed": {
				NameKey:        "room.tutorial.shed.name",
				DescriptionKey: "room.tutorial.shed.description",
				Name:           "器械棚",
				Description:    "角落里堆着一些练习器械。",
			},
		},
		items: map[ItemID]Item{
			"item.tutorial.north": {
				NameKey:        "item.tutorial.north.name",
				InnerName:      "north",
				DescriptionKey: "item.tutorial.north.description",
				Name:           "北方",
				Description:    "北方通向练习场。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.yard"}},
					{DefinitionID: "tag.lockable", Params: map[string]any{"key_item_id": "item.tutorial.old_lantern"}},
				},
			},
			"item.tutorial.south": {
				NameKey:        "item.tutorial.south.name",
				InnerName:      "south",
				DescriptionKey: "item.tutorial.south.description",
				Name:           "南方",
				Description:    "南方通向练习场入口。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "south", "target": "room.tutorial.start"}},
				},
			},
			"item.tutorial.northeast": {
				NameKey:        "item.tutorial.northeast.name",
				InnerName:      "northeast",
				DescriptionKey: "item.tutorial.northeast.description",
				Name:           "东北方",
				Description:    "东北方通向器械棚。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "northeast", "target": "room.tutorial.shed"}},
				},
			},
			"item.tutorial.east": {
				NameKey:        "item.tutorial.east.name",
				InnerName:      "east",
				DescriptionKey: "item.tutorial.east.description",
				Name:           "东方",
				Description:    "东方通向器械棚。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "east", "target": "room.tutorial.shed"}},
				},
			},
			"item.tutorial.west": {
				NameKey:        "item.tutorial.west.name",
				InnerName:      "west",
				DescriptionKey: "item.tutorial.west.description",
				Name:           "西方",
				Description:    "西方通向练习场。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "west", "target": "room.tutorial.yard"}},
				},
			},
			"item.tutorial.southwest": {
				NameKey:        "item.tutorial.southwest.name",
				InnerName:      "southwest",
				DescriptionKey: "item.tutorial.southwest.description",
				Name:           "西南方",
				Description:    "西南方通向练习场入口。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "southwest", "target": "room.tutorial.start"}},
				},
			},
			"item.tutorial.old_lantern": {
				NameKey:        "item.tutorial.old_lantern.name",
				InnerName:      "old lantern",
				DescriptionKey: "item.tutorial.old_lantern.description",
				Name:           "旧油灯",
				Description:    "灯罩上蒙着一层灰，里面还剩一点灯油。",
				Aliases:        []string{"jiuyoudeng", "old_lantern"},
				Tags: []TagInstance{
					{DefinitionID: "tag.carryable", Params: map[string]any{}},
					{DefinitionID: "tag.lightable", Params: map[string]any{}},
				},
			},
			"item.tutorial.practice_sword": {
				NameKey:        "item.tutorial.practice_sword.name",
				InnerName:      "practice sword",
				DescriptionKey: "item.tutorial.practice_sword.description",
				Name:           "练习木剑",
				Description:    "一把被许多人握过的木剑，剑柄已经磨得发亮。",
				Aliases:        []string{"lianximujian"},
				Tags:           []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
			},
		},
		itemLocations: map[ItemID]ItemLocation{
			"item.tutorial.north":     RoomItemLocation{RoomID: "room.tutorial.start"},
			"item.tutorial.northeast": RoomItemLocation{RoomID: "room.tutorial.start"},
			"item.tutorial.south":     RoomItemLocation{RoomID: "room.tutorial.yard"},
			"item.tutorial.east":      RoomItemLocation{RoomID: "room.tutorial.yard"},
			"item.tutorial.west":      RoomItemLocation{RoomID: "room.tutorial.shed"},
			"item.tutorial.southwest": RoomItemLocation{RoomID: "room.tutorial.shed"},
			"item.tutorial.old_lantern": RoomItemLocation{
				RoomID: "room.tutorial.start",
			},
			"item.tutorial.practice_sword": RoomItemLocation{
				RoomID: "room.tutorial.yard",
			},
		},
		progressionDefinitions: tutorialProgressionDefinitions(),
		players:                make(map[PlayerID]PlayerEntity),
		tagDefinitions:         make(map[TagID]TagDefinition),
		contentVerbs:           make(map[string]VerbEntry),
	}
	initBuiltinTags(w)
	return w
}

func initBuiltinTags(w *World) {
	for _, def := range builtinTagDefs() {
		w.RegisterTag(def)
	}
}

func NewFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) *World {
	rooms := make(map[RoomID]Room, len(snapshot.Rooms))
	for roomID := range snapshot.Rooms {
		nameKey := catalog.RoomNames[roomID]
		descriptionKey := catalog.RoomDescriptions[roomID]
		rooms[RoomID(roomID)] = Room{
			NameKey:        string(nameKey),
			DescriptionKey: string(descriptionKey),
			Name:           catalog.Text[nameKey],
			Description:    catalog.Text[descriptionKey],
		}
	}

	items := make(map[ItemID]Item, len(snapshot.Items))
	for itemID := range snapshot.Items {
		serverItem := snapshot.Items[itemID]
		nameKey := catalog.ItemDisplayNames[itemID]
		descriptionKey := catalog.ItemDescriptions[itemID]
		items[ItemID(itemID)] = Item{
			NameKey:        string(nameKey),
			InnerName:      catalog.Text[serverItem.InnerNameKey],
			DescriptionKey: string(descriptionKey),
			Name:           catalog.Text[nameKey],
			Description:    catalog.Text[descriptionKey],
			Aliases:        textKeysToStrings(catalog, serverItem.Aliases),
			Tags:           worldTags(serverItem.Tags),
		}
	}

	itemLocations := make(map[ItemID]ItemLocation, len(snapshot.ItemLocations))
	for itemID, roomID := range snapshot.ItemLocations {
		itemLocations[ItemID(itemID)] = RoomItemLocation{RoomID: RoomID(roomID)}
	}

	w := &World{
		startRoom:              RoomID(snapshot.StartRoomID),
		rooms:                  rooms,
		items:                  items,
		itemLocations:          itemLocations,
		progressionDefinitions: progressionDefinitionsFromSnapshot(snapshot, catalog),
		players:                make(map[PlayerID]PlayerEntity),
		tagDefinitions:         make(map[TagID]TagDefinition),
		contentVerbs:           make(map[string]VerbEntry, len(snapshot.Verbs)),
	}
	for verbID, sv := range snapshot.Verbs {
		w.contentVerbs[string(verbID)] = VerbEntry{
			Name:       string(verbID),
			Source:     VerbSourceContent,
			MessageKey: string(sv.MessageKey),
		}
	}
	initBuiltinTags(w)
	return w
}

func worldTags(tags []content.ServerTag) []TagInstance {
	result := make([]TagInstance, 0, len(tags))
	for _, tag := range tags {
		if tag.Carryable {
			result = append(result, TagInstance{
				DefinitionID: "tag.carryable",
				Params:       map[string]any{},
			})
			continue
		}
		if tag.Lightable {
			result = append(result, TagInstance{
				DefinitionID: "tag.lightable",
				Params:       map[string]any{},
			})
			continue
		}
		if tag.Container != nil {
			result = append(result, TagInstance{
				DefinitionID: "tag.container",
				Params:       map[string]any{"capacity": tag.Container.Capacity},
			})
			continue
		}
		if tag.Lockable != nil {
			result = append(result, TagInstance{
				DefinitionID: "tag.lockable",
				Params:       map[string]any{"key_item_id": string(tag.Lockable.KeyItemID)},
			})
			continue
		}
		if tag.Exit == nil {
			continue
		}
		params := map[string]any{"target": string(tag.Exit.TargetRoomID)}
		if tag.Exit.Direction != "" {
			params["direction"] = string(tag.Exit.Direction)
		}
		result = append(result, TagInstance{
			DefinitionID: "tag.exit",
			Params:       params,
		})
	}
	return result
}

func (w *World) ProgressionDefinitions() progression.Definitions {
	return w.progressionDefinitions
}

func progressionDefinitionsFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) progression.Definitions {
	questIDs := make([]content.QuestID, 0, len(snapshot.Quests))
	for questID := range snapshot.Quests {
		questIDs = append(questIDs, questID)
	}
	slices.Sort(questIDs)
	if len(questIDs) == 0 {
		return progression.Definitions{}
	}
	questID := questIDs[0]
	quest := snapshot.Quests[questID]
	definitions := progression.Definitions{
		Quest: progression.QuestDefinition{
			ID:       string(questID),
			Name:     catalog.Text[quest.NameKey],
			StageIDs: questStageIDsToStrings(quest.StageIDs),
		},
		Stages: make(map[string]progression.StageDefinition, len(snapshot.QuestStages)),
	}
	for stageID, stage := range snapshot.QuestStages {
		definitions.Stages[string(stageID)] = progression.StageDefinition{
			ID:         string(stageID),
			Text:       catalog.Text[stage.TextKey],
			Conditions: progressionConditions(stage.FinishConditions),
			NextID:     string(stage.NextStageID),
		}
	}
	return definitions
}

func questStageIDsToStrings(ids []content.QuestStageID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, string(id))
	}
	return values
}

func progressionConditions(conditions []content.ServerQuestCondition) []progression.ConditionDefinition {
	values := make([]progression.ConditionDefinition, 0, len(conditions))
	for _, condition := range conditions {
		values = append(values, progression.ConditionDefinition{
			Kind:   string(condition.Kind),
			ItemID: string(condition.ItemID),
			RoomID: string(condition.RoomID),
			Text:   conditionText(condition),
		})
	}
	return values
}

func conditionText(condition content.ServerQuestCondition) string {
	switch condition.Kind {
	case content.QuestConditionGotItem:
		return "获取旧油灯"
	case content.QuestConditionMovedRoom:
		return "到达练习场"
	case content.QuestConditionExaminedItem:
		return "查看练习木剑"
	default:
		return string(condition.Kind)
	}
}

func tutorialProgressionDefinitions() progression.Definitions {
	return progression.Definitions{
		Quest: progression.QuestDefinition{
			ID:   "quest.tutorial.first_steps",
			Name: "教程任务",
			StageIDs: []string{
				"quest.tutorial.first_steps.stage.get_lantern",
				"quest.tutorial.first_steps.stage.enter_yard",
				"quest.tutorial.first_steps.stage.examine_sword",
			},
		},
		Stages: map[string]progression.StageDefinition{
			"quest.tutorial.first_steps.stage.get_lantern": {
				ID:     "quest.tutorial.first_steps.stage.get_lantern",
				Text:   "拿起旧油灯。",
				NextID: "quest.tutorial.first_steps.stage.enter_yard",
		Conditions: []progression.ConditionDefinition{
				{Kind: string(progression.TriggerGotItem), ItemID: "item.tutorial.old_lantern", Text: "获取旧油灯"},
			},
			},
			"quest.tutorial.first_steps.stage.enter_yard": {
				ID:     "quest.tutorial.first_steps.stage.enter_yard",
				Text:   "前往练习场。",
				NextID: "quest.tutorial.first_steps.stage.examine_sword",
				Conditions: []progression.ConditionDefinition{
				{Kind: string(progression.TriggerMovedRoom), RoomID: "room.tutorial.yard", Text: "到达练习场"},
			},
		},
		"quest.tutorial.first_steps.stage.examine_sword": {
			ID:   "quest.tutorial.first_steps.stage.examine_sword",
			Text: "查看练习木剑。",
			Conditions: []progression.ConditionDefinition{
				{Kind: string(progression.TriggerExaminedItem), ItemID: "item.tutorial.practice_sword", Text: "查看练习木剑"},
				},
			},
		},
	}
}

func textKeysToStrings(catalog content.ClientCatalog, keys []content.TextKey) []string {
	if len(keys) == 0 {
		return nil
	}
	values := make([]string, 0, len(keys))
	for _, key := range keys {
		if value, ok := catalog.Text[key]; ok {
			values = append(values, value)
		}
	}
	return values
}
