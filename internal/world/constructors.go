package world

import (
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
			"room.tutorial.weight_room": {
				NameKey: "room.tutorial.weight_room.name", DescriptionKey: "room.tutorial.weight_room.description",
				Name: "测重房", Description: "一间摆满各种砝码和重物的房间，墙上贴着\"负重测试\"的牌子。",
			},
			"room.tutorial.dark_cave": {
				NameKey: "room.tutorial.dark_cave.name", DescriptionKey: "room.tutorial.dark_cave.description",
				Name: "黑暗洞穴", Description: "伸手不见五指的黑暗包围着你，什么也看不见。",
				Dark: true,
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
				Weight: 5, Volume: 2,
			},
			"item.tutorial.practice_sword": {
				NameKey: "item.tutorial.practice_sword.name", InnerName: "practice sword", DescriptionKey: "item.tutorial.practice_sword.description",
				Name: "练习木剑", Description: "一把被许多人握过的木剑，剑柄已经磨得发亮。",
				Aliases: []string{"lianximujian"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight:  3, Volume: 3,
			},
			"item.tutorial.training_relic": {
				NameKey: "item.tutorial.training_relic.name", InnerName: "training relic", DescriptionKey: "item.tutorial.training_relic.description",
				Name: "练功徽章", Description: "一枚金属徽章，上面刻着\"练功有成\"四个字。",
				Aliases: []string{"liangonghuizhang", "training_relic"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight:  1, Volume: 1,
			},
			// 测重房出口
			"item.yard.east": {
				NameKey: "item.yard.east.name", InnerName: "east", DescriptionKey: "item.yard.east.description",
				Name: "东方通道", Description: "一条向东延伸的通道，通向测重房。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "east", "target": "room.tutorial.weight_room"}}},
			},
			"item.weight_room.west": {
				NameKey: "item.weight_room.west.name", InnerName: "west", DescriptionKey: "item.weight_room.west.description",
				Name: "西方出口", Description: "向西返回物品庭院。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "west", "target": "room.tutorial.item_yard"}}},
			},
			"item.weight_room.north": {
				NameKey: "item.weight_room.north.name", InnerName: "north", DescriptionKey: "item.weight_room.north.description",
				Name: "北方洞口", Description: "一个幽深的洞口，里面一片漆黑。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "north", "target": "room.tutorial.dark_cave"}}},
			},
			"item.dark_cave.south": {
				NameKey: "item.dark_cave.south.name", InnerName: "south", DescriptionKey: "item.dark_cave.south.description",
				Name: "南方出口", Description: "一丝光亮从南方透进来。",
				Tags: []TagInstance{{DefinitionID: "tag.exit", Params: map[string]any{"direction": "south", "target": "room.tutorial.weight_room"}}},
			},
			"item.tutorial.treasure_chest": {
				NameKey: "item.tutorial.treasure_chest.name", InnerName: "treasure chest", DescriptionKey: "item.tutorial.treasure_chest.description",
				Name: "宝藏箱", Description: "一个小巧的宝箱，里面装满了宝石。",
				Tags:   []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight: 2, Volume: 1,
			},
			// 测重房物品
			"item.tutorial.small_weight": {
				NameKey: "item.tutorial.small_weight.name", InnerName: "small weight", DescriptionKey: "item.tutorial.small_weight.description",
				Name: "小砝码", Description: "一个很小的砝码，几乎感觉不到重量。",
				Tags:   []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight: 1, Volume: 1,
			},
			"item.tutorial.medium_weight": {
				NameKey: "item.tutorial.medium_weight.name", InnerName: "medium weight", DescriptionKey: "item.tutorial.medium_weight.description",
				Name: "中砝码", Description: "一个中等大小的砝码。",
				Tags:   []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight: 5, Volume: 2,
			},
			"item.tutorial.large_weight": {
				NameKey: "item.tutorial.large_weight.name", InnerName: "large weight", DescriptionKey: "item.tutorial.large_weight.description",
				Name: "大砝码", Description: "一个沉重的砝码，拎起来有些费劲。",
				Tags:   []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight: 10, Volume: 5,
			},
			"item.tutorial.iron_ingot": {
				NameKey: "item.tutorial.iron_ingot.name", InnerName: "iron ingot", DescriptionKey: "item.tutorial.iron_ingot.description",
				Name: "铁锭", Description: "一块沉甸甸的铁锭，体积不大却很重。",
				Aliases: []string{"tieding"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight:  8, Volume: 1,
			},
			"item.tutorial.cotton_bale": {
				NameKey: "item.tutorial.cotton_bale.name", InnerName: "cotton bale", DescriptionKey: "item.tutorial.cotton_bale.description",
				Name: "棉花包", Description: "一大包棉花，占地方但很轻。",
				Aliases: []string{"mianhuabao"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight:  1, Volume: 8,
			},
			"item.tutorial.barbell": {
				NameKey: "item.tutorial.barbell.name", InnerName: "barbell", DescriptionKey: "item.tutorial.barbell.description",
				Name: "杠铃", Description: "一副沉重的杠铃，一般人举不动。",
				Aliases: []string{"gangling"},
				Tags:    []TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
				Weight:  15, Volume: 5,
			},
			"item.tutorial.boulder": {
				NameKey: "item.tutorial.boulder.name", InnerName: "boulder", DescriptionKey: "item.tutorial.boulder.description",
				Name: "巨石", Description: "一块嵌在地面的巨石，纹丝不动。",
				Weight: 999, Volume: 999,
			},
			"item.tutorial.storage_pouch": {
				NameKey: "item.tutorial.storage_pouch.name", InnerName: "storage pouch", DescriptionKey: "item.tutorial.storage_pouch.description",
				Name: "收纳袋", Description: "一个轻便的收纳袋，可以把东西装进去减少占用的空间。",
				Tags: []TagInstance{
					{DefinitionID: "tag.carryable", Params: map[string]any{}},
					{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
				},
				Weight: 2, Volume: 1,
			},
		},
		itemLocations: map[ItemID]ItemLocation{
			"item.hall.north":              RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.hall.east":               RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.hall.portal":             RoomItemLocation{RoomID: "room.tutorial.hall"},
			"item.yard.south":              RoomItemLocation{RoomID: "room.tutorial.item_yard"},
			"item.lock_hall.west":          RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.lock_hall.east":          RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.lock_chamber.west":       RoomItemLocation{RoomID: "room.tutorial.lock_chamber"},
			"item.quest_start.portal":      RoomItemLocation{RoomID: "room.tutorial.quest_start"},
			"item.tutorial.old_lantern":    RoomItemLocation{RoomID: "room.tutorial.lock_hall"},
			"item.tutorial.practice_sword": RoomItemLocation{RoomID: "room.tutorial.item_yard"},
			"item.tutorial.training_relic": RoomItemLocation{RoomID: "room.tutorial.lock_chamber"},
			"item.yard.east":               RoomItemLocation{RoomID: "room.tutorial.item_yard"},
			"item.weight_room.west":        RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.weight_room.north":       RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.dark_cave.south":         RoomItemLocation{RoomID: "room.tutorial.dark_cave"},
			"item.tutorial.treasure_chest": RoomItemLocation{RoomID: "room.tutorial.dark_cave"},
			"item.tutorial.small_weight":   RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.medium_weight":  RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.large_weight":   RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.iron_ingot":     RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.cotton_bale":    RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.barbell":        RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.boulder":        RoomItemLocation{RoomID: "room.tutorial.weight_room"},
			"item.tutorial.storage_pouch":  RoomItemLocation{RoomID: "room.tutorial.weight_room"},
		},
		progressionDefinitions: tutorialProgressionDefinitions(),
		players:                make(map[PlayerID]PlayerEntity),
		tagDefinitions:         make(map[TagID]TagDefinition),
		contentVerbs:           make(map[string]VerbEntry),
		trackedQuests:          make(map[PlayerID]string),
		containerOpen:          make(map[ItemID]bool),
		litItems:               make(map[ItemID]bool),
	}
	initBuiltinTags(w)
	return w
}

func initBuiltinTags(w *World) {
	for _, def := range builtinTagDefs() {
		w.tagDefinitions[def.ID] = def
	}
}

func NewFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) *World {
	rooms := make(map[RoomID]Room, len(snapshot.Rooms))
	for roomID, sRoom := range snapshot.Rooms {
		nameKey := catalog.RoomNames[roomID]
		descriptionKey := catalog.RoomDescriptions[roomID]
		rooms[RoomID(roomID)] = Room{
			NameKey:        string(nameKey),
			DescriptionKey: string(descriptionKey),
			Name:           catalog.Text[nameKey],
			Description:    catalog.Text[descriptionKey],
			Dark:           sRoom.Dark,
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
			Parts:          worldParts(serverItem.Parts),
			Weight:         serverItem.Weight,
			Volume:         serverItem.Volume,
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
		trackedQuests:          make(map[PlayerID]string),
		containerOpen:          make(map[ItemID]bool),
		litItems:               make(map[ItemID]bool),
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

func worldParts(parts map[content.PartID]content.ServerPart) map[string]ItemPart {
	if len(parts) == 0 {
		return nil
	}
	result := make(map[string]ItemPart, len(parts))
	for id, part := range parts {
		result[string(id)] = ItemPart{
			Tags: worldTags(part.Tags),
		}
	}
	return result
}

func (w *World) ProgressionDefinitions() progression.Definitions {
	return w.progressionDefinitions
}

func progressionDefinitionsFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) progression.Definitions {
	defs := progression.Definitions{
		Quests: make(map[string]progression.QuestDefinition, len(snapshot.Quests)),
		Stages: make(map[string]progression.StageDefinition, len(snapshot.QuestStages)),
	}
	for questID, quest := range snapshot.Quests {
		defs.Quests[string(questID)] = progression.QuestDefinition{
			ID:                   string(questID),
			Name:                 catalog.Text[quest.NameKey],
			StageIDs:             questStageIDsToStrings(quest.StageIDs),
			Activation:           progression.ActivationMode(quest.Activation),
			ActivationConditions: progressionConditions(quest.ActivationConditions),
			Repeatable:           quest.Repeatable,
		}
	}
	for stageID, stage := range snapshot.QuestStages {
		defs.Stages[string(stageID)] = progression.StageDefinition{
			ID:         string(stageID),
			Text:       catalog.Text[stage.TextKey],
			Conditions: progressionConditions(stage.FinishConditions),
			NextID:     string(stage.NextStageID),
		}
	}
	return defs
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
		if condition.ItemID != "" {
			return "获取 " + string(condition.ItemID)
		}
		return "获取物品"
	case content.QuestConditionMovedRoom:
		if condition.RoomID != "" {
			return "到达 " + string(condition.RoomID)
		}
		return "到达指定地点"
	case content.QuestConditionExaminedItem:
		if condition.ItemID != "" {
			return "查看 " + string(condition.ItemID)
		}
		return "查看物品"
	default:
		return string(condition.Kind)
	}
}

func tutorialProgressionDefinitions() progression.Definitions {
	return progression.Definitions{
		Quests: map[string]progression.QuestDefinition{
			"quest.tutorial.first_steps": {
				ID:   "quest.tutorial.first_steps",
				Name: "教程任务",
				StageIDs: []string{
					"quest.tutorial.first_steps.stage.get_lantern",
					"quest.tutorial.first_steps.stage.enter_chamber",
					"quest.tutorial.first_steps.stage.examine_relic",
				},
			},
			"quest.tutorial.practice_sword": {
				ID:   "quest.tutorial.practice_sword",
				Name: "练剑任务",
				StageIDs: []string{
					"quest.tutorial.practice_sword.stage.get_sword",
					"quest.tutorial.practice_sword.stage.examine_sword",
				},
			},
		},
		Stages: map[string]progression.StageDefinition{
			"quest.tutorial.first_steps.stage.get_lantern": {
				ID:     "quest.tutorial.first_steps.stage.get_lantern",
				Text:   "拿起旧油灯。",
				NextID: "quest.tutorial.first_steps.stage.enter_chamber",
				Conditions: []progression.ConditionDefinition{
					{Kind: string(progression.TriggerGotItem), ItemID: "item.tutorial.old_lantern", Text: "获取旧油灯"},
				},
			},
			"quest.tutorial.first_steps.stage.enter_chamber": {
				ID:     "quest.tutorial.first_steps.stage.enter_chamber",
				Text:   "进入密室。",
				NextID: "quest.tutorial.first_steps.stage.examine_relic",
				Conditions: []progression.ConditionDefinition{
					{Kind: string(progression.TriggerMovedRoom), RoomID: "room.tutorial.lock_chamber", Text: "进入密室"},
				},
			},
			"quest.tutorial.first_steps.stage.examine_relic": {
				ID:   "quest.tutorial.first_steps.stage.examine_relic",
				Text: "查看练功徽章。",
				Conditions: []progression.ConditionDefinition{
					{Kind: string(progression.TriggerExaminedItem), ItemID: "item.tutorial.training_relic", Text: "查看练功徽章"},
				},
			},
			"quest.tutorial.practice_sword.stage.get_sword": {
				ID:     "quest.tutorial.practice_sword.stage.get_sword",
				Text:   "拿起练习木剑。",
				NextID: "quest.tutorial.practice_sword.stage.examine_sword",
				Conditions: []progression.ConditionDefinition{
					{Kind: string(progression.TriggerGotItem), ItemID: "item.tutorial.practice_sword", Text: "获取练习木剑"},
				},
			},
			"quest.tutorial.practice_sword.stage.examine_sword": {
				ID:   "quest.tutorial.practice_sword.stage.examine_sword",
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
