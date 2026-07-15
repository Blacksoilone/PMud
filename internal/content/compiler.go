package content

import "maps"

func Compile(source ContentSource) (CompiledContent, error) {
	server := ServerSnapshot{
		StartRoomID:   source.StartRoomID,
		Rooms:         make(map[RoomID]ServerRoom, len(source.Rooms)),
		Items:         make(map[ItemID]ServerItem, len(source.Items)),
		ItemLocations: make(map[ItemID]RoomID, len(source.Items)),
	}
	client := ClientCatalog{
		RoomNames:        make(map[RoomID]TextKey, len(source.Rooms)),
		RoomDescriptions: make(map[RoomID]TextKey, len(source.Rooms)),
		ItemDisplayNames: make(map[ItemID]TextKey, len(source.Items)),
		ItemInnerNames:   make(map[ItemID]TextKey, len(source.Items)),
		ItemDescriptions: make(map[ItemID]TextKey, len(source.Items)),
		ItemAliases:      make(map[ItemID][]TextKey, len(source.Items)),
		Text:             make(map[TextKey]string, len(source.Text)),
	}

	for _, room := range source.Rooms {
		exits := make(map[Direction]RoomID, len(room.Exits))
		maps.Copy(exits, room.Exits)
		server.Rooms[room.ID] = ServerRoom{Exits: exits}
		client.RoomNames[room.ID] = room.NameKey
		client.RoomDescriptions[room.ID] = room.DescriptionKey
	}

	for _, item := range source.Items {
		server.Items[item.ID] = ServerItem{}
		server.ItemLocations[item.ID] = item.InitialRoom
		client.ItemDisplayNames[item.ID] = item.DisplayNameKey
		client.ItemInnerNames[item.ID] = item.InnerNameKey
		client.ItemDescriptions[item.ID] = item.DescriptionKey
		if len(item.Aliases) > 0 {
			client.ItemAliases[item.ID] = append([]TextKey(nil), item.Aliases...)
		}
	}

	maps.Copy(client.Text, source.Text)

	return CompiledContent{
		Server: server,
		Client: client,
	}, nil
}
