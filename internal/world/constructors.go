package world

import (
	"slices"

	"PMud/internal/content"
	"PMud/internal/progression"
)

func New() *World {
	w := &World{
		startRoom: "room.tutorial.hall",
		rooms: map[RoomID]Room{
			"room.tutorial.hall": {
				NameKey: "room.tutorial.hall.name", DescriptionKey: "room.tutorial.hall.description",
				Name: "教学大厅", Description: "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。",
			},
			"room.tutorial.item_yard": {
				NameKey: "room.tutorial.item_yard.name", DescriptionKey: "room.tutorial.item_yard.description",
				Name: "物品庭院", Description: "庭院里散落着各种练习用品，地上还有几把木剑。",
			},
			"room.tutorial.lock_hall": {
				NameKey: "room.tutorial.lock_hall.name", DescriptionKey: "room.tutorial.lock_hall.description",
				Name: "锁钥厅", Description: "墙壁上挂着一盏旧油灯，东侧有一扇厚重的铁门。",
			},
			"room.tutorial.lock_chamber": {
				NameKey: "room.tutorial.lock_chamber.name", DescriptionKey: "room.tutorial.lock_chamber.description",
				Name: "密室", Description: "一间昏暗的小室，空气中有股陈旧的灰尘味。",
			},
			"room.tutorial.quest_start": {
				NameKey: "room.tutorial.quest_start.name", DescriptionKey: "room.tutorial.quest_start.description",
				Name: "任务之门", Description: "你感到一股神秘的力量在这里流动。",
			},
		},
		items: map[ItemID]Item{
			// 大厅出口
			"item.hall.north": {
				NameKey: "item.hall.north.name", InnerName: "north", DescriptionKey: "item.hall.north.description",
				Name: "北方通路", Description: "一条向北延伸的石板路。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.item_yard"}}},
			},
			"item.hall.east": {
				NameKey: "item.hall.east.name", InnerName: "east", DescriptionKey: "item.hall.east.description",
				Name: "东方通路", Description: "一条向东的走廊。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "east", "target": "room.tutorial.lock_hall"}}},
			},
			"item.hall.portal": {
				NameKey: "item.hall.portal.name", InnerName: "portal", DescriptionKey: "item.hall.portal.description",
				Name: "传送门", Description: "一扇泛着蓝光的传送门。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "portal", "target": "room.tutorial.quest_start"}}},
			},
			// 物品庭院出口
			"item.yard.south": {
				NameKey: "item.yard.south.name", InnerName: "south", DescriptionKey: "item.yard.south.description",
				Name: "南方通路", Description: "一条向南的石板路，通向大厅。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "south", "target": "room.tutorial.hall"}}},
			},
			// 锁钥厅出口
			"item.lock_hall.west": {
				NameKey: "item.lock_hall.west.name", InnerName: "west", DescriptionKey: "item.lock_hall.west.description",
				Name: "西方通路", Description: "向西返回大厅的通道。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "west", "target": "room.tutorial.hall"}}},
			},
			"item.lock_hall.east": {
				NameKey: "item.lock_hall.east.name", InnerName: "east", DescriptionKey: "item.lock_hall.east.description",
				Name: "铁门", Description: "一扇厚重的铁门，似乎需要什么才能打开。",
				Tags: []TagInstance{
					{DefinitionID: "tag.exit", Params: map[string]any{"direction": "east", "target": "room.tutorial.lock_chamber"}},
					{DefinitionID: "tag.lockable", Params: map[string]any{"key_item_id": "item.tutorial.old_lantern"}},
				},
			},
			// 密室出口
			"item.lock_chamber.west": {
				NameKey: "item.lock_chamber.west.name", InnerName: "west", DescriptionKey: "item.lock_chamber.west.description",
				Name: "西方出口", Description: "向西返回锁钥厅。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "west", "target": "room.tutorial.lock_hall"}}},
			},
			// 任务起点出口
			"item.quest_start.portal": {
				NameKey: "item.quest_start.portal.name", InnerName: "portal", DescriptionKey: "item.quest_start.portal.description",
				Name: "传送门", Description: "一扇泛着蓝光的传送门。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "portal", "target": "room.tutorial.hall"}}},
			},
			// 游戏物品
			"item.tutorial.old_lantern": {
				NameKey: "item.tutorial.old_lantern.name", InnerName: "old lantern", DescriptionKey: "item.tutorial.old_lantern.description",
				Name: "旧油灯", Description: "灯罩上蒙着一层灰，里面还剩一点灯油。",
				Aliases: []string{"jiuyoudeng", "old_lantern"},
				Tags: []TagInstance{
					{DefinitionID: "tag.carryable", Params: map[string]any{}},
					{DefinitionID: "tag.lightable", Params: map[string]any{}},
				},
			},
			"item.tutorial.practice_sword": {
				NameKey: "item.tutorial.practice_sword.name", InnerName: "practice sword", DescriptionKey: "item.tutorial.practice_sword.description",
				Name: "练习木剑", Description: "一把被许多人握过的木剑，剑柄已经磨得发亮。",
				Aliases: []string{"lianximujian"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
			},
		},
		itemLocations: map[ItemID]ItemLocation{
			"item.hall.north":        RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.hall.east":         RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.hall.portal":       RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.yard.south":        RoomItemLocation{RoomID: "room.tutorial.item_yard"},
			"item.lock_hall.west":    RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.lock_hall.east":    RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.lock_chamber.west": RoomItemLocation{RoomID: "room.tutorial.lock_chamber"},
			"item.quest_start.portal":  RoomItemLocation{RoomID: "room.tutorial.quest_start"},
			"item.tutorial.old_lantern": RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.tutorial.practice_sword": RoomItemLocation{RoomID: "room.tutorial.item_yard"},
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
		return "到达物品庭院"
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
				Text:   "前往物品庭院。",
				NextID: "quest.tutorial.first_steps.stage.examine_sword",
				Conditions: []progression.ConditionDefinition{
				{Kind: string(progression.TriggerMovedRoom), RoomID: "room.tutorial.item_yard", Text: "到达物品庭院"},
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
