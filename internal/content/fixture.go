package content

func TutorialSource() ContentSource {
	return ContentSource{
		StartRoomID: "room.tutorial.start",
		Rooms: []RoomSource{
			{
				ID:             "room.tutorial.start",
				NameKey:        "room.tutorial.start.name",
				DescriptionKey: "room.tutorial.start.description",
				Exits: map[Direction]RoomID{
					"north": "room.tutorial.yard",
				},
			},
			{
				ID:             "room.tutorial.yard",
				NameKey:        "room.tutorial.yard.name",
				DescriptionKey: "room.tutorial.yard.description",
				Exits: map[Direction]RoomID{
					"south": "room.tutorial.start",
				},
			},
		},
		Items: []ItemSource{
			{
				ID:             "item.tutorial.old_lantern",
				DisplayNameKey: "item.tutorial.old_lantern.name",
				InnerNameKey:   "item.tutorial.old_lantern.inner_name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				Aliases: []TextKey{
					"item.tutorial.old_lantern.alias.jiuyoudeng",
					"item.tutorial.old_lantern.alias.old_lantern",
				},
				InitialRoom: "room.tutorial.start",
			},
			{
				ID:             "item.tutorial.practice_sword",
				DisplayNameKey: "item.tutorial.practice_sword.name",
				InnerNameKey:   "item.tutorial.practice_sword.inner_name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				InitialRoom:    "room.tutorial.yard",
			},
		},
		Text: map[TextKey]string{
			"room.tutorial.start.name":                    "练习场入口",
			"room.tutorial.start.description":             "这里是练习场的入口。北边传来木剑碰撞的声音。",
			"room.tutorial.yard.name":                     "练习场",
			"room.tutorial.yard.description":              "几根木桩立在泥地上，地面满是被踩出的脚印。",
			"item.tutorial.old_lantern.name":              "旧油灯",
			"item.tutorial.old_lantern.inner_name":        "old lantern",
			"item.tutorial.old_lantern.description":       "灯罩上蒙着一层灰，里面还剩一点灯油。",
			"item.tutorial.old_lantern.alias.jiuyoudeng":  "jiuyoudeng",
			"item.tutorial.old_lantern.alias.old_lantern": "old_lantern",
			"item.tutorial.practice_sword.name":           "练习木剑",
			"item.tutorial.practice_sword.inner_name":     "practice sword",
			"item.tutorial.practice_sword.description":    "一把被许多人握过的木剑，剑柄已经磨得发亮。",
			"system.empty_input":                          "你没有输入任何内容",
			"system.help":                                 "可用命令: look/l, go <direction>, north/n, south/s, east/e, west/w, up/u, down/d, northeast/ne, northwest/nw, southeast/se, southwest/sw, get/take <item>, drop <item>, examine/x/inspect <item>, inventory/i, help\n方向: north/n/北, south/s/南, east/e, west/w, up/u, down/d, northeast/ne, northwest/nw, southeast/se, southwest/sw",
			"system.move.blocked":                         "你不能往那个方向走。",
			"system.item.not_here":                        "这里没有那个东西。",
			"system.item.taken":                           "你拿起了{item}。",
			"system.item.not_carried":                     "你没有那个东西。",
			"system.item.dropped":                         "你放下了{item}。",
			"system.unknown_command":                      "未知命令: {input}",
			"system.room.missing":                         "你迷失在不存在的地方。",
		},
	}
}
