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
				ID:          "item.tutorial.old_lantern",
				NameKey:     "item.tutorial.old_lantern.name",
				InitialRoom: "room.tutorial.start",
			},
			{
				ID:          "item.tutorial.practice_sword",
				NameKey:     "item.tutorial.practice_sword.name",
				InitialRoom: "room.tutorial.yard",
			},
		},
		Text: map[TextKey]string{
			"room.tutorial.start.name":          "练习场入口",
			"room.tutorial.start.description":   "这里是练习场的入口。北边传来木剑碰撞的声音。",
			"room.tutorial.yard.name":           "练习场",
			"room.tutorial.yard.description":    "几根木桩立在泥地上，地面满是被踩出的脚印。",
			"item.tutorial.old_lantern.name":    "旧油灯",
			"item.tutorial.practice_sword.name": "练习木剑",
		},
	}
}
