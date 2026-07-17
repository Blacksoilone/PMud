package content

import (
	"fmt"
	"maps"
)

func Compile(source ContentSource) (CompiledContent, error) {
	server := ServerSnapshot{
		StartRoomID:   source.StartRoomID,
		Rooms:         make(map[RoomID]ServerRoom, len(source.Rooms)),
		Items:         make(map[ItemID]ServerItem, len(source.Items)),
		ItemLocations: make(map[ItemID]RoomID, len(source.Items)),
		Quests:        make(map[QuestID]ServerQuest, len(source.Quests)),
		QuestStages:   make(map[QuestStageID]ServerQuestStage, len(source.QuestStages)),
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
		server.Rooms[room.ID] = ServerRoom{}
		client.RoomNames[room.ID] = room.NameKey
		client.RoomDescriptions[room.ID] = room.DescriptionKey
	}

	directionsByRoom := make(map[RoomID]map[Direction]ItemID)
	for _, item := range source.Items {
		tags, err := compileTags(item, source.Rooms, source.Text)
		if err != nil {
			return CompiledContent{}, err
		}
		server.Items[item.ID] = ServerItem{
			DisplayNameKey: item.DisplayNameKey,
			InnerNameKey:   item.InnerNameKey,
			DescriptionKey: item.DescriptionKey,
			Aliases:        append([]TextKey(nil), item.Aliases...),
			Tags:           tags,
		}
		for _, tag := range tags {
			if tag.Exit == nil || tag.Exit.Direction == "" {
				continue
			}
			roomDirections := directionsByRoom[item.InitialRoom]
			if roomDirections == nil {
				roomDirections = make(map[Direction]ItemID)
				directionsByRoom[item.InitialRoom] = roomDirections
			}
			if previous, exists := roomDirections[tag.Exit.Direction]; exists {
				return CompiledContent{}, fmt.Errorf("room %q: duplicate exit direction %q on items %q and %q", item.InitialRoom, tag.Exit.Direction, previous, item.ID)
			}
			roomDirections[tag.Exit.Direction] = item.ID
		}
		server.ItemLocations[item.ID] = item.InitialRoom
		client.ItemDisplayNames[item.ID] = item.DisplayNameKey
		client.ItemInnerNames[item.ID] = item.InnerNameKey
		client.ItemDescriptions[item.ID] = item.DescriptionKey
		if len(item.Aliases) > 0 {
			client.ItemAliases[item.ID] = append([]TextKey(nil), item.Aliases...)
		}
	}

	for _, quest := range source.Quests {
		server.Quests[quest.ID] = ServerQuest{
			NameKey:  quest.NameKey,
			StageIDs: append([]QuestStageID(nil), quest.StageIDs...),
		}
	}

	for _, stage := range source.QuestStages {
		conditions := make([]ServerQuestCondition, 0, len(stage.FinishConditions))
		for _, condition := range stage.FinishConditions {
			conditions = append(conditions, ServerQuestCondition(condition))
		}
		server.QuestStages[stage.ID] = ServerQuestStage{
			TextKey:          stage.TextKey,
			FinishConditions: conditions,
			NextStageID:      stage.NextStageID,
		}
	}

	maps.Copy(client.Text, source.Text)

	return CompiledContent{
		Server: server,
		Client: client,
	}, nil
}

func compileTags(item ItemSource, rooms []RoomSource, text map[TextKey]string) ([]ServerTag, error) {
	result := make([]ServerTag, 0, len(item.Tags))
	for _, tag := range item.Tags {
		switch tag.ID {
		case TagCarryable:
			if len(tag.Params) != 0 {
				return nil, fmt.Errorf("item %q: carryable tag accepts no parameters", item.ID)
			}
			result = append(result, ServerTag{Carryable: true})
			continue
		case TagExit:
		default:
			return nil, fmt.Errorf("item %q: unknown tag %q", item.ID, tag.ID)
		}
		target, ok := tag.Params["target_room_id"]
		if !ok || target == "" {
			return nil, fmt.Errorf("item %q: exit tag requires target_room_id", item.ID)
		}
		if !hasRoom(rooms, RoomID(target)) {
			return nil, fmt.Errorf("item %q: exit target room %q does not exist", item.ID, target)
		}
		result = append(result, ServerTag{Exit: &ExitTag{
			Direction:    inferDirection(text[item.InnerNameKey]),
			TargetRoomID: RoomID(target),
		}})
	}
	return result, nil
}

func hasRoom(rooms []RoomSource, id RoomID) bool {
	for _, room := range rooms {
		if room.ID == id {
			return true
		}
	}
	return false
}

func inferDirection(innerName string) Direction {
	switch innerName {
	case "north", "northeast", "east", "southeast", "south", "southwest", "west", "northwest":
		return Direction(innerName)
	default:
		return ""
	}
}
