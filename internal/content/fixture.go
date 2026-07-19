package content

func TutorialSource() ContentSource {
	return ContentSource{
		StartRoomID: "room.tutorial.hall",
		Verbs: []VerbSource{
			{ID: "light", MessageKey: "verb.light.default"},
		},
		Rooms: []RoomSource{
			{ID: "room.tutorial.hall", NameKey: "room.tutorial.hall.name", DescriptionKey: "room.tutorial.hall.description"},
			{ID: "room.tutorial.item_yard", NameKey: "room.tutorial.item_yard.name", DescriptionKey: "room.tutorial.item_yard.description"},
			{ID: "room.tutorial.lock_hall", NameKey: "room.tutorial.lock_hall.name", DescriptionKey: "room.tutorial.lock_hall.description"},
			{ID: "room.tutorial.lock_chamber", NameKey: "room.tutorial.lock_chamber.name", DescriptionKey: "room.tutorial.lock_chamber.description"},
			{ID: "room.tutorial.quest_start", NameKey: "room.tutorial.quest_start.name", DescriptionKey: "room.tutorial.quest_start.description"},
		},
		Items: []ItemSource{
			// === 大厅出口 ===
			{
				ID: "item.hall.north", DisplayNameKey: "item.hall.north.name", InnerNameKey: "item.hall.north.inner_name",
				DescriptionKey: "item.hall.north.description", InitialRoom: "room.tutorial.hall",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.item_yard"}}},
			},
			{
				ID: "item.hall.east", DisplayNameKey: "item.hall.east.name", InnerNameKey: "item.hall.east.inner_name",
				DescriptionKey: "item.hall.east.description", InitialRoom: "room.tutorial.hall",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.lock_hall"}}},
			},
			{
				ID: "item.hall.portal", DisplayNameKey: "item.hall.portal.name", InnerNameKey: "item.hall.portal.inner_name",
				DescriptionKey: "item.hall.portal.description", InitialRoom: "room.tutorial.hall",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.quest_start", "direction": "portal"}}},
			},
			// === 物品庭院出口 ===
			{
				ID: "item.yard.south", DisplayNameKey: "item.yard.south.name", InnerNameKey: "item.yard.south.inner_name",
				DescriptionKey: "item.yard.south.description", InitialRoom: "room.tutorial.item_yard",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.hall"}}},
			},
			// === 锁钥厅出口 ===
			{
				ID: "item.lock_hall.west", DisplayNameKey: "item.lock_hall.west.name", InnerNameKey: "item.lock_hall.west.inner_name",
				DescriptionKey: "item.lock_hall.west.description", InitialRoom: "room.tutorial.lock_hall",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.hall"}}},
			},
			{
				ID: "item.lock_hall.east", DisplayNameKey: "item.lock_hall.east.name", InnerNameKey: "item.lock_hall.east.inner_name",
				DescriptionKey: "item.lock_hall.east.description", InitialRoom: "room.tutorial.lock_hall",
				Tags: []SourceTag{
					{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.lock_chamber"}},
					{ID: TagLockable, Params: map[string]string{"key_item_id": "item.tutorial.old_lantern"}},
				},
			},
			// === 密室出口 ===
			{
				ID: "item.lock_chamber.west", DisplayNameKey: "item.lock_chamber.west.name", InnerNameKey: "item.lock_chamber.west.inner_name",
				DescriptionKey: "item.lock_chamber.west.description", InitialRoom: "room.tutorial.lock_chamber",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.lock_hall"}}},
			},
			// === 任务起点出口 ===
			{
				ID: "item.quest_start.portal", DisplayNameKey: "item.quest_start.portal.name", InnerNameKey: "item.quest_start.portal.inner_name",
				DescriptionKey: "item.quest_start.portal.description", InitialRoom: "room.tutorial.quest_start",
				Tags: []SourceTag{{ID: TagExit, Params: map[string]string{"target_room_id": "room.tutorial.hall", "direction": "portal"}}},
			},
			// === 游戏物品 ===
			{
				ID:             "item.tutorial.old_lantern",
				DisplayNameKey: "item.tutorial.old_lantern.name",
				InnerNameKey:   "item.tutorial.old_lantern.inner_name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				Aliases: []TextKey{
					"item.tutorial.old_lantern.alias.jiuyoudeng",
					"item.tutorial.old_lantern.alias.old_lantern",
				},
				InitialRoom: "room.tutorial.lock_hall",
				Tags:        []SourceTag{{ID: TagCarryable}, {ID: TagLightable}},
			},
			{
				ID:             "item.tutorial.practice_sword",
				DisplayNameKey: "item.tutorial.practice_sword.name",
				InnerNameKey:   "item.tutorial.practice_sword.inner_name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				Aliases:        []TextKey{"item.tutorial.practice_sword.alias.lianximujian"},
				InitialRoom:    "room.tutorial.item_yard",
				Tags:           []SourceTag{{ID: TagCarryable}},
			},
			{
				ID:             "item.tutorial.training_relic",
				DisplayNameKey: "item.tutorial.training_relic.name",
				InnerNameKey:   "item.tutorial.training_relic.inner_name",
				DescriptionKey: "item.tutorial.training_relic.description",
				Aliases: []TextKey{
					"item.tutorial.training_relic.alias.liangonghuizhang",
					"item.tutorial.training_relic.alias.training_relic",
				},
				InitialRoom: "room.tutorial.lock_chamber",
				Tags:        []SourceTag{{ID: TagCarryable}},
			},
		},
		Quests: []QuestSource{
			{
				ID:      "quest.tutorial.first_steps",
				NameKey: "quest.tutorial.first_steps.name",
				StageIDs: []QuestStageID{
					"quest.tutorial.first_steps.stage.get_lantern",
					"quest.tutorial.first_steps.stage.enter_chamber",
					"quest.tutorial.first_steps.stage.examine_relic",
				},
			},
		},
		QuestStages: []QuestStageSource{
			{
				ID:      "quest.tutorial.first_steps.stage.get_lantern",
				TextKey: "quest.tutorial.first_steps.stage.get_lantern.text",
				FinishConditions: []QuestConditionSource{
					{Kind: QuestConditionGotItem, ItemID: "item.tutorial.old_lantern"},
				},
				NextStageID: "quest.tutorial.first_steps.stage.enter_chamber",
			},
			{
				ID:      "quest.tutorial.first_steps.stage.enter_chamber",
				TextKey: "quest.tutorial.first_steps.stage.enter_chamber.text",
				FinishConditions: []QuestConditionSource{
					{Kind: QuestConditionMovedRoom, RoomID: "room.tutorial.lock_chamber"},
				},
				NextStageID: "quest.tutorial.first_steps.stage.examine_relic",
			},
			{
				ID:      "quest.tutorial.first_steps.stage.examine_relic",
				TextKey: "quest.tutorial.first_steps.stage.examine_relic.text",
				FinishConditions: []QuestConditionSource{
					{Kind: QuestConditionExaminedItem, ItemID: "item.tutorial.training_relic"},
				},
			},
		},
		Text: map[TextKey]string{
			// 房间名称与描述
			"room.tutorial.hall.name":              "教学大厅",
			"room.tutorial.hall.description":       "大厅宽敞明亮，四周墙壁上挂着几幅地图。这里连通着多个区域。",
			"room.tutorial.item_yard.name":         "物品庭院",
			"room.tutorial.item_yard.description":  "庭院里散落着各种练习用品，地上还有几把木剑。",
			"room.tutorial.lock_hall.name":         "锁钥厅",
			"room.tutorial.lock_hall.description":  "墙壁上挂着一盏旧油灯，东侧有一扇厚重的铁门。",
			"room.tutorial.lock_chamber.name":      "密室",
			"room.tutorial.lock_chamber.description": "一间昏暗的小室，空气中有股陈旧的灰尘味。",
			"room.tutorial.quest_start.name":       "任务之门",
			"room.tutorial.quest_start.description": "你感到一股神秘的力量在这里流动。",

			// 出口物品
			"item.hall.north.name":          "北方通路",
			"item.hall.north.inner_name":    "north",
			"item.hall.north.description":   "一条向北延伸的石板路。",
			"item.hall.east.name":           "东方通路",
			"item.hall.east.inner_name":     "east",
			"item.hall.east.description":    "一条向东的走廊。",
			"item.hall.portal.name":         "传送门",
			"item.hall.portal.inner_name":   "portal",
			"item.hall.portal.description":  "一扇泛着蓝光的传送门。",
			"item.yard.south.name":          "南方通路",
			"item.yard.south.inner_name":    "south",
			"item.yard.south.description":   "一条向南的石板路，通向大厅。",
			"item.lock_hall.west.name":      "西方通路",
			"item.lock_hall.west.inner_name": "west",
			"item.lock_hall.west.description": "向西返回大厅的通道。",
			"item.lock_hall.east.name":      "铁门",
			"item.lock_hall.east.inner_name": "east",
			"item.lock_hall.east.description": "一扇厚重的铁门，似乎需要什么才能打开。",
			"item.lock_chamber.west.name":      "西方出口",
			"item.lock_chamber.west.inner_name": "west",
			"item.lock_chamber.west.description": "向西返回锁钥厅。",
			"item.quest_start.portal.name":      "传送门",
			"item.quest_start.portal.inner_name": "portal",
			"item.quest_start.portal.description": "一扇泛着蓝光的传送门。",

			// 游戏物品
			"item.tutorial.old_lantern.name":                  "旧油灯",
			"item.tutorial.old_lantern.inner_name":            "old lantern",
			"item.tutorial.old_lantern.description":           "灯罩上蒙着一层灰，里面还剩一点灯油。",
			"item.tutorial.old_lantern.alias.jiuyoudeng":      "jiuyoudeng",
			"item.tutorial.old_lantern.alias.old_lantern":     "old_lantern",
			"item.tutorial.practice_sword.name":               "练习木剑",
			"item.tutorial.practice_sword.inner_name":         "practice sword",
			"item.tutorial.practice_sword.description":        "一把被许多人握过的木剑，剑柄已经磨得发亮。",
			"item.tutorial.practice_sword.alias.lianximujian": "lianximujian",
			"item.tutorial.training_relic.name":                  "练功徽章",
			"item.tutorial.training_relic.inner_name":            "training relic",
			"item.tutorial.training_relic.description":           "一枚金属徽章，上面刻着\"练功有成\"四个字。",
			"item.tutorial.training_relic.alias.liangonghuizhang": "liangonghuizhang",
			"item.tutorial.training_relic.alias.training_relic":  "training_relic",

			// 任务
			"quest.tutorial.first_steps.name":                       "教程任务",
			"quest.tutorial.first_steps.stage.get_lantern.text":     "拿起旧油灯。",
			"quest.tutorial.first_steps.stage.enter_chamber.text":   "进入密室。",
			"quest.tutorial.first_steps.stage.examine_relic.text":   "查看练功徽章。",

			// 系统消息
			"system.empty_input":             "你没有输入任何内容",
			"system.help": "可用命令: look/l, go <direction>, north/n, south/s, east/e, west/w, up/u, down/d, northeast/ne, northwest/nw, southeast/se, southwest/sw, get/take <item>, drop <item>, examine/x/inspect <item>, inventory/i, quest, verb/verbs, help\n方向: north/n/北, south/s/南, east/e, west/w, up/u, down/d, northeast/ne, northwest/nw, southeast/se, southwest/sw",
			"system.internal_error":          "服务器内部错误，请稍后重试。",
			"system.look.observed":           "你重新观察了周围。",
			"system.move.blocked":            "你不能往那个方向走。",
			"system.move.locked":             "门锁着。",
			"system.item.not_here":           "这里没有那个东西。",
			"system.item.taken":              "你拿起了{item}。",
			"system.item.not_carried":        "你没有那个东西。",
			"system.item.dropped":            "你放下了{item}。",
			"system.quest.progress":          "任务更新: {state}",
			"system.unknown_command":         "未知命令: {input}",
			"system.room.missing":            "你迷失在不存在的地方。",

			// 内容动词
			"verb.light.default": "你点亮了{item}。",
		},
	}
}
