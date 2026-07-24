package world

import (
	"PMud/internal/content"
	"PMud/internal/progression"
)

func New() *World {
	w := initWorld()

	addRoom := func(id, nameKey, descKey, name, desc string, dark bool, tags ...TagInstance) {
		w.store.Add(&Entity{
			ID: id, NameKey: nameKey, DescriptionKey: descKey,
			Name: name, Description: desc, Tags: tags,
			Room: &RoomData{Dark: dark},
		})
	}
	}
	addRoom("room.tutorial.hall", "room.tutorial.hall.name", "room.tutorial.hall.description",
		"教学大厅", "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。", false)
	addRoom("room.tutorial.item_yard", "room.tutorial.item_yard.name", "room.tutorial.item_yard.description",
		"物品庭院", "庭院里散落着各种练习用品，地上还有几把木剑。", false)
	addRoom("room.tutorial.lock_hall", "room.tutorial.lock_hall.name", "room.tutorial.lock_hall.description",
		"锁钥厅", "墙壁上挂着一盏旧油灯，东侧有一扇厚重的铁门。", false)
	addRoom("room.tutorial.lock_chamber", "room.tutorial.lock_chamber.name", "room.tutorial.lock_chamber.description",
		"密室", "一间昏暗的小室，空气中有股陈旧的灰尘味。", false)
	addRoom("room.tutorial.quest_start", "room.tutorial.quest_start.name", "room.tutorial.quest_start.description",
		"任务之门", "你感到一股神秘的力量在这里流动。", false)
	addRoom("room.tutorial.weight_room", "room.tutorial.weight_room.name", "room.tutorial.weight_room.description",
		"测重房", "一间摆满各种砝码和重物的房间，墙上贴着\"负重测试\"的牌子。", false)
	addRoom("room.tutorial.dark_cave", "room.tutorial.dark_cave.name", "room.tutorial.dark_cave.description",
		"黑暗洞穴", "伸手不见五指的黑暗包围着你，什么也看不见。", true)

	addExit := func(id, nameKey, name, desc string, tags []TagInstance, direction, target string) {
		w.store.Add(&Entity{
			ID: id, NameKey: nameKey, Name: name, Description: desc, Tags: tags,
			Exit: &ExitData{Direction: direction, TargetRoomID: target},
		})
	}
	addExit("exit:room.hall.north", "item.hall.north.name", "北方通路", "一条向北延伸的石板路。",
		nil, "north", "room.tutorial.item_yard")
	addExit("exit:room.hall.east", "item.hall.east.name", "东方通路", "一条向东的走廊。",
		nil, "east", "room.tutorial.lock_hall")
	addExit("exit:room.hall.portal", "item.hall.portal.name", "传送门", "一扇泛着蓝光的传送门。",
		nil, "portal", "room.tutorial.quest_start")
	addExit("exit:room.yard.south", "item.yard.south.name", "南方通路", "一条向南的石板路，通向大厅。",
		nil, "south", "room.tutorial.hall")
	addExit("exit:room.lock_hall.west", "item.lock_hall.west.name", "西方通路", "向西返回大厅的通道。",
		nil, "west", "room.tutorial.hall")
	addExit("exit:room.lock_hall.east", "item.lock_hall.east.name", "铁门", "一扇厚重的铁门，似乎需要什么才能打开。",
		[]TagInstance{
			{DefinitionID: "tag.lockable", Params: map[string]any{"key_item_id": "item.tutorial.old_lantern"}},
		}, "east", "room.tutorial.lock_chamber")
	addExit("exit:room.lock_chamber.west", "item.lock_chamber.west.name", "西方出口", "向西返回锁钥厅。",
		nil, "west", "room.tutorial.lock_hall")
	addExit("exit:room.quest_start.portal", "item.quest_start.portal.name", "传送门", "一扇泛着蓝光的传送门。",
		nil, "portal", "room.tutorial.hall")
	addExit("exit:room.yard.east", "item.yard.east.name", "东方通道", "一条向东延伸的通道，通向测重房。",
		nil, "east", "room.tutorial.weight_room")
	addExit("exit:room.weight_room.west", "item.weight_room.west.name", "西方出口", "向西返回物品庭院。",
		nil, "west", "room.tutorial.item_yard")
	addExit("exit:room.weight_room.north", "item.weight_room.north.name", "北方洞口", "一个幽深的洞口，里面一片漆黑。",
		nil, "north", "room.tutorial.dark_cave")
	addExit("exit:room.dark_cave.south", "item.dark_cave.south.name", "南方出口", "一丝光亮从南方透进来。",
		nil, "south", "room.tutorial.weight_room")

	place := func(entityID, roomID string) { w.store.PlaceInRoom(entityID, roomID) }
	place("exit:room.hall.north", "room.tutorial.hall")
	place("exit:room.hall.east", "room.tutorial.hall")
	place("exit:room.hall.portal", "room.tutorial.hall")
	place("exit:room.yard.south", "room.tutorial.item_yard")
	place("exit:room.lock_hall.west", "room.tutorial.lock_hall")
	place("exit:room.lock_hall.east", "room.tutorial.lock_hall")
	place("exit:room.lock_chamber.west", "room.tutorial.lock_chamber")
	place("exit:room.quest_start.portal", "room.tutorial.quest_start")
	place("exit:room.yard.east", "room.tutorial.item_yard")
	place("exit:room.weight_room.west", "room.tutorial.weight_room")
	place("exit:room.weight_room.north", "room.tutorial.weight_room")
	place("exit:room.dark_cave.south", "room.tutorial.dark_cave")

	addItem := func(id, nameKey, innerName, descKey, name, desc string, wgt, vol int, tags []TagInstance, aliases []string) {
		w.store.Add(&Entity{
			ID: id, InnerName: innerName, NameKey: nameKey, DescriptionKey: descKey,
			Name: name, Description: desc, Tags: tags, Aliases: aliases,
			Item: &ItemData{Weight: wgt, Volume: vol},
		})
	}
	addItem("item.tutorial.old_lantern", "item.tutorial.old_lantern.name", "old lantern",
		"item.tutorial.old_lantern.description", "旧油灯", "灯罩上蒙着一层灰，里面还剩一点灯油。",
		5, 2,
		[]TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.lightable", Params: map[string]any{}},
		},
		[]string{"jiuyoudeng", "old_lantern"})
	addItem("item.tutorial.practice_sword", "item.tutorial.practice_sword.name", "practice sword",
		"item.tutorial.practice_sword.description", "练习木剑", "一把被许多人握过的木剑，剑柄已经磨得发亮。",
		3, 3,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		[]string{"lianximujian"})
	addItem("item.tutorial.training_relic", "item.tutorial.training_relic.name", "training relic",
		"item.tutorial.training_relic.description", "练功徽章", "一枚金属徽章，上面刻着\"练功有成\"四个字。",
		1, 1,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		[]string{"liangonghuizhang", "training_relic"})
	addItem("item.tutorial.treasure_chest", "item.tutorial.treasure_chest.name", "treasure chest",
		"item.tutorial.treasure_chest.description", "宝藏箱", "一个小巧的宝箱，里面装满了宝石。",
		2, 1,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}}, nil)
	addItem("item.tutorial.small_weight", "item.tutorial.small_weight.name", "small weight",
		"item.tutorial.small_weight.description", "小砝码", "一个很小的砝码，几乎感觉不到重量。",
		1, 1,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}}, nil)
	addItem("item.tutorial.medium_weight", "item.tutorial.medium_weight.name", "medium weight",
		"item.tutorial.medium_weight.description", "中砝码", "一个中等大小的砝码。",
		5, 2,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}}, nil)
	addItem("item.tutorial.large_weight", "item.tutorial.large_weight.name", "large weight",
		"item.tutorial.large_weight.description", "大砝码", "一个沉重的砝码，拎起来有些费劲。",
		10, 5,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}}, nil)
	addItem("item.tutorial.iron_ingot", "item.tutorial.iron_ingot.name", "iron ingot",
		"item.tutorial.iron_ingot.description", "铁锭", "一块沉甸甸的铁锭，体积不大却很重。",
		8, 1,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		[]string{"tieding"})
	addItem("item.tutorial.cotton_bale", "item.tutorial.cotton_bale.name", "cotton bale",
		"item.tutorial.cotton_bale.description", "棉花包", "一大包棉花，占地方但很轻。",
		1, 8,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		[]string{"mianhuabao"})
	addItem("item.tutorial.barbell", "item.tutorial.barbell.name", "barbell",
		"item.tutorial.barbell.description", "杠铃", "一副沉重的杠铃，一般人举不动。",
		15, 5,
		[]TagInstance{{DefinitionID: "tag.carryable", Params: map[string]any{}}},
		[]string{"gangling"})
	addItem("item.tutorial.boulder", "item.tutorial.boulder.name", "boulder",
		"item.tutorial.boulder.description", "巨石", "一块嵌在地面的巨石，纹丝不动。",
		999, 999, nil, nil)
	addItem("item.tutorial.storage_pouch", "item.tutorial.storage_pouch.name", "storage pouch",
		"item.tutorial.storage_pouch.description", "收纳袋", "一个轻便的收纳袋，可以把东西装进去减少占用的空间。",
		2, 1,
		[]TagInstance{
			{DefinitionID: "tag.carryable", Params: map[string]any{}},
			{DefinitionID: "tag.container", Params: map[string]any{"capacity": 5}},
		}, nil)

	place("item.tutorial.old_lantern", "room.tutorial.lock_hall")
	place("item.tutorial.practice_sword", "room.tutorial.item_yard")
	place("item.tutorial.training_relic", "room.tutorial.lock_chamber")
	place("item.tutorial.treasure_chest", "room.tutorial.dark_cave")
	place("item.tutorial.small_weight", "room.tutorial.weight_room")
	place("item.tutorial.medium_weight", "room.tutorial.weight_room")
	place("item.tutorial.large_weight", "room.tutorial.weight_room")
	place("item.tutorial.iron_ingot", "room.tutorial.weight_room")
	place("item.tutorial.cotton_bale", "room.tutorial.weight_room")
	place("item.tutorial.barbell", "room.tutorial.weight_room")
	place("item.tutorial.boulder", "room.tutorial.weight_room")
	place("item.tutorial.storage_pouch", "room.tutorial.weight_room")

	initBuiltinTags(w)
	return w
}

func initWorld() *World {
	return &World{
		store:                  NewEntityStore(),
		startRoom:              "room.tutorial.hall",
		progressionDefinitions: tutorialProgressionDefinitions(),
		tagDefinitions:         make(map[TagID]TagDefinition),
		contentVerbs:           make(map[string]VerbEntry),
		containerContents:      make(map[string][]EntityID),
		containerOpen:          make(map[EntityID]bool),
		litItems:               make(map[EntityID]bool),
		trackedQuests:          make(map[PlayerID]string),
	}
}

func initBuiltinTags(w *World) {
	for _, def := range builtinTagDefs() {
		w.tagDefinitions[def.ID] = def
	}
}

func NewFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) *World {
	w := &World{
		store:                  NewEntityStore(),
		startRoom:              EntityID(snapshot.StartRoomID),
		progressionDefinitions: progressionDefinitionsFromSnapshot(snapshot, catalog),
		tagDefinitions:         make(map[TagID]TagDefinition),
		contentVerbs:           make(map[string]VerbEntry, len(snapshot.Verbs)),
		containerContents:      make(map[string][]EntityID),
		containerOpen:          make(map[EntityID]bool),
		litItems:               make(map[EntityID]bool),
		trackedQuests:          make(map[PlayerID]string),
	}

	for roomID, sRoom := range snapshot.Rooms {
		nameKey := catalog.RoomNames[roomID]
		descriptionKey := catalog.RoomDescriptions[roomID]
		w.store.Add(&Entity{
			ID: EntityID(roomID), NameKey: string(nameKey), DescriptionKey: string(descriptionKey),
			Name: catalog.Text[nameKey], Description: catalog.Text[descriptionKey],
			Room: &RoomData{Dark: sRoom.Dark},
		})
	}

	// 物品与出口
	for itemID, serverItem := range snapshot.Items {
		nameKey := catalog.ItemDisplayNames[itemID]
		descriptionKey := catalog.ItemDescriptions[itemID]
		entityID := EntityID(itemID)
		tags := worldTags(serverItem.Tags)
		innerNameKey := catalog.ItemInnerNames[itemID]
		innerName := catalog.Text[innerNameKey]

		// 从 serverItem.Tags 中提取出口信息（worldTags 跳过了 exit tag）
		var exitTarget, exitDir string
		for _, t := range serverItem.Tags {
			if t.Exit != nil {
				exitTarget = string(t.Exit.TargetRoomID)
				exitDir = string(t.Exit.Direction)
				break
			}
		}
		if exitTarget != "" {
			w.store.Add(&Entity{
				ID: entityID, InnerName: innerName, NameKey: string(nameKey), DescriptionKey: string(descriptionKey),
				Name: catalog.Text[nameKey], Description: catalog.Text[descriptionKey],
				Exit: &ExitData{Direction: exitDir, TargetRoomID: EntityID(exitTarget)},
			})
		} else {
			w.store.Add(&Entity{
				ID: entityID, InnerName: innerName, NameKey: string(nameKey), DescriptionKey: string(descriptionKey),
				Name: catalog.Text[nameKey], Description: catalog.Text[descriptionKey],
				Aliases: textKeysToStrings(catalog, serverItem.Aliases),
				Tags:    tags,
				Parts:   worldParts(serverItem.Parts),
				Item:    &ItemData{Weight: serverItem.Weight, Volume: serverItem.Volume},
			})
		}
	}

	// 物品位置
	for itemID, roomID := range snapshot.ItemLocations {
		if ent := w.store.Get(EntityID(itemID)); ent != nil {
			w.store.PlaceInRoom(EntityID(itemID), EntityID(roomID))
		}
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

// extractExitFromTags 检查 tags 中是否含 tag.exit，如果有则提取参数并从 tags 中删除。
// 返回目标房间和方向（方向可能为空）。
func extractExitFromTags(tags []TagInstance) (target, direction string) {
	for i, t := range tags {
		if t.DefinitionID == "tag.exit" {
			target, _ = t.Params["target"].(string)
			direction, _ = t.Params["direction"].(string)
			// 从切片中移除该 tag
			tags[i] = tags[len(tags)-1]
			_ = tags[:len(tags)-1]
			return
		}
	}
	return "", ""
}

func worldTags(tags []content.ServerTag) []TagInstance {
	result := make([]TagInstance, 0, len(tags))
	for _, tag := range tags {
		if tag.Exit != nil {
			// tag.exit 由 extractExitFromTags 在 NewFromSnapshot 中处理，这里跳过
			continue
		}
		if tag.Carryable {
			result = append(result, TagInstance{DefinitionID: "tag.carryable", Params: map[string]any{}})
			continue
		}
		if tag.Lightable {
			result = append(result, TagInstance{DefinitionID: "tag.lightable", Params: map[string]any{}})
			continue
		}
		if tag.Container != nil {
			result = append(result, TagInstance{DefinitionID: "tag.container", Params: map[string]any{"capacity": tag.Container.Capacity}})
			continue
		}
		if tag.Lockable != nil {
			result = append(result, TagInstance{DefinitionID: "tag.lockable", Params: map[string]any{"key_item_id": string(tag.Lockable.KeyItemID)}})
			continue
		}
		if tag.GenericID != "" {
			params := make(map[string]any, len(tag.GenericParams))
			for k, v := range tag.GenericParams {
				params[k] = v
			}
			result = append(result, TagInstance{DefinitionID: TagID(tag.GenericID), Params: params})
		}
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
