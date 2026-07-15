package world

import "PMud/internal/content"

func New() *World {
	return &World{
		startRoom: "room.tutorial.start",
		rooms: map[RoomID]Room{
			"room.tutorial.start": {
				NameKey:        "room.tutorial.start.name",
				DescriptionKey: "room.tutorial.start.description",
				Name:           "练习场入口",
				Description:    "这里是练习场的入口。北边传来木剑碰撞的声音。",
				Exits: map[string]RoomID{
					"north": "room.tutorial.yard",
				},
			},
			"room.tutorial.yard": {
				NameKey:        "room.tutorial.yard.name",
				DescriptionKey: "room.tutorial.yard.description",
				Name:           "练习场",
				Description:    "几根木桩立在泥地上，地面满是被踩出的脚印。",
				Exits: map[string]RoomID{
					"south": "room.tutorial.start",
				},
			},
		},
		items: map[ItemID]Item{
			"item.tutorial.old_lantern": {
				NameKey:        "item.tutorial.old_lantern.name",
				DescriptionKey: "item.tutorial.old_lantern.description",
				Name:           "旧油灯",
				Description:    "灯罩上蒙着一层灰，里面还剩一点灯油。",
			},
			"item.tutorial.practice_sword": {
				NameKey:        "item.tutorial.practice_sword.name",
				DescriptionKey: "item.tutorial.practice_sword.description",
				Name:           "练习木剑",
				Description:    "一把被许多人握过的木剑，剑柄已经磨得发亮。",
			},
		},
		itemLocations: map[ItemID]ItemLocation{
			"item.tutorial.old_lantern": RoomItemLocation{
				RoomID: "room.tutorial.start",
			},
			"item.tutorial.practice_sword": RoomItemLocation{
				RoomID: "room.tutorial.yard",
			},
		},
	}
}

func NewFromSnapshot(snapshot content.ServerSnapshot, catalog content.ClientCatalog) *World {
	rooms := make(map[RoomID]Room, len(snapshot.Rooms))
	for roomID, room := range snapshot.Rooms {
		exits := make(map[string]RoomID, len(room.Exits))
		for direction, targetRoomID := range room.Exits {
			exits[string(direction)] = RoomID(targetRoomID)
		}

		nameKey := catalog.RoomNames[roomID]
		descriptionKey := catalog.RoomDescriptions[roomID]
		rooms[RoomID(roomID)] = Room{
			NameKey:        string(nameKey),
			DescriptionKey: string(descriptionKey),
			Name:           catalog.Text[nameKey],
			Description:    catalog.Text[descriptionKey],
			Exits:          exits,
		}
	}

	items := make(map[ItemID]Item, len(snapshot.Items))
	for itemID := range snapshot.Items {
		nameKey := catalog.ItemDisplayNames[itemID]
		descriptionKey := catalog.ItemDescriptions[itemID]
		items[ItemID(itemID)] = Item{
			NameKey:        string(nameKey),
			DescriptionKey: string(descriptionKey),
			Name:           catalog.Text[nameKey],
			Description:    catalog.Text[descriptionKey],
		}
	}

	itemLocations := make(map[ItemID]ItemLocation, len(snapshot.ItemLocations))
	for itemID, roomID := range snapshot.ItemLocations {
		itemLocations[ItemID(itemID)] = RoomItemLocation{RoomID: RoomID(roomID)}
	}

	return &World{
		startRoom:     RoomID(snapshot.StartRoomID),
		rooms:         rooms,
		items:         items,
		itemLocations: itemLocations,
	}
}
