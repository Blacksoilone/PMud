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
		Verbs:         make(map[VerbID]ServerVerb, len(source.Verbs)),
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
		VerbNames:        make(map[VerbID]TextKey, len(source.Verbs)),
		Text:             make(map[TextKey]string, len(source.Text)),
	}

	for _, room := range source.Rooms {
		server.Rooms[room.ID] = ServerRoom{Dark: room.Dark}
		client.RoomNames[room.ID] = room.NameKey
		client.RoomDescriptions[room.ID] = room.DescriptionKey
	}

	directionsByRoom := make(map[RoomID]map[Direction]ItemID)
	for _, item := range source.Items {
		tags, err := compileTags(item, source.Rooms, source.Text)
		if err != nil {
			return CompiledContent{}, err
		}
		parts, err := compileParts(item, source.Text)
		if err != nil {
			return CompiledContent{}, err
		}
		server.Items[item.ID] = ServerItem{
			DisplayNameKey: item.DisplayNameKey,
			InnerNameKey:   item.InnerNameKey,
			DescriptionKey: item.DescriptionKey,
			Aliases:        append([]TextKey(nil), item.Aliases...),
			Tags:           tags,
			Parts:          parts,
			Weight:         item.Weight,
			Volume:         item.Volume,
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

	for _, verb := range source.Verbs {
		server.Verbs[verb.ID] = ServerVerb{MessageKey: verb.MessageKey}
		if verb.MessageKey != "" {
			client.VerbNames[verb.ID] = verb.MessageKey
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
		compiled, err := compileOneTag(item.ID, tag, rooms, text[item.InnerNameKey])
		if err != nil {
			return nil, err
		}
		result = append(result, compiled)
	}
	return result, nil
}

func compileOneTag(itemID ItemID, tag SourceTag, rooms []RoomSource, innerName string) (ServerTag, error) {
	switch tag.ID {
	case TagCarryable:
		if len(tag.Params) != 0 {
			return ServerTag{}, fmt.Errorf("item %q: carryable tag accepts no parameters", itemID)
		}
		return ServerTag{Carryable: true}, nil
	case TagLightable:
		if len(tag.Params) != 0 {
			return ServerTag{}, fmt.Errorf("item %q: lightable tag accepts no parameters", itemID)
		}
		return ServerTag{Lightable: true}, nil
	case TagContainer:
		capacity := 1
		if capStr, ok := tag.Params["capacity"]; ok {
			if _, err := fmt.Sscanf(capStr, "%d", &capacity); err != nil {
				return ServerTag{}, fmt.Errorf("item %q: container tag capacity must be an integer, got %q", itemID, capStr)
			}
		}
		return ServerTag{Container: &ContainerTag{Capacity: capacity}}, nil
	case TagLockable:
		keyID, ok := tag.Params["key_item_id"]
		if !ok || keyID == "" {
			return ServerTag{}, fmt.Errorf("item %q: lockable tag requires key_item_id", itemID)
		}
		return ServerTag{Lockable: &LockableTag{KeyItemID: ItemID(keyID)}}, nil
	case TagExit:
		target, ok := tag.Params["target_room_id"]
		if !ok || target == "" {
			return ServerTag{}, fmt.Errorf("item %q: exit tag requires target_room_id", itemID)
		}
		if !hasRoom(rooms, RoomID(target)) {
			return ServerTag{}, fmt.Errorf("item %q: exit target room %q does not exist", itemID, target)
		}
		return ServerTag{Exit: &ExitTag{
			Direction:    inferDirection(innerName),
			TargetRoomID: RoomID(target),
		}}, nil
	default:
		return ServerTag{
			GenericID:     tag.ID,
			GenericParams: tag.Params,
		}, nil
	}
}

func compileParts(item ItemSource, text map[TextKey]string) (map[PartID]ServerPart, error) {
	if len(item.Parts) == 0 {
		return nil, nil
	}
	result := make(map[PartID]ServerPart, len(item.Parts))
	for partID, part := range item.Parts {
		tags := make([]ServerTag, 0, len(part.SourceTags))
		for _, srcTag := range part.SourceTags {
			compiled, err := compileOneTag(item.ID, srcTag, nil, text[item.InnerNameKey])
			if err != nil {
				return nil, fmt.Errorf("item %q part %q: %v", item.ID, partID, err)
			}
			tags = append(tags, compiled)
		}
		result[partID] = ServerPart{Tags: tags}
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
