package presentation

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

type TextRenderer struct{}

func (TextRenderer) Render(event Event) string {
	switch e := event.(type) {
	case SystemMessageEvent:
		if e.MessageKey == "" {
			return line("system", field("message", e.Message))
		}
		fields := make([]string, 0, len(e.Fields)+1)
		fields = append(fields, field("message_key", e.MessageKey))
		keys := make([]string, 0, len(e.Fields))
		for key := range e.Fields {
			keys = append(keys, key)
		}
		slices.Sort(keys)
		for _, key := range keys {
			fields = append(fields, field(key, e.Fields[key]))
		}
		return line("system", fields...)
	case RoomObservationEvent:
		neighbors := make([]string, 0, len(e.Neighbors))
		for direction, room := range e.Neighbors {
			neighbors = append(neighbors, direction+"="+room)
		}
		slices.Sort(neighbors)
		return line(
			"room",
			field("room", e.Room),
			field("name_key", e.NameKey),
			field("description_key", e.DescriptionKey),
			field("exits", strings.Join(e.Exits, ",")),
			optionalField("neighbors", strings.Join(neighbors, ",")),
			field("items", strings.Join(e.Items, ",")),
		)
	case InventoryEvent:
		wt := strconv.Itoa(e.WeightCurrent) + "/" + strconv.Itoa(e.WeightMax)
		vol := strconv.Itoa(e.VolumeCurrent) + "/" + strconv.Itoa(e.VolumeMax)
		return line("inventory",
			field("items", strings.Join(e.Items, ",")),
			field("weight", wt),
			field("volume", vol),
		)
	case QuestStatusEvent:
		return line(
			"quest",
			field("quest_id", e.QuestID),
			field("quest_name", e.QuestName),
			field("stage_id", e.StageID),
			field("stage_text", e.StageText),
			field("conditions", strings.Join(e.Conditions, ",")),
			field("state", e.State),
		)
	case QuestListEvent:
		type questListItem struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Stage string `json:"stage"`
			State string `json:"state"`
		}
		items := make([]questListItem, 0, len(e.Quests))
		for _, q := range e.Quests {
			items = append(items, questListItem{ID: q.QuestID, Name: q.QuestName, Stage: q.StageText, State: q.State})
		}
		encoded, err := json.Marshal(items)
		if err != nil {
			return line("quest_list", field("count", strconv.Itoa(len(e.Quests))))
		}
		return line(
			"quest_list",
			field("count", strconv.Itoa(len(e.Quests))),
			field("items", string(encoded)),
			optionalField("tracked", e.Tracked),
		)
	case ItemObservationEvent:
		tags := strings.Join(e.Tags, ",")
		partStrs := make([]string, 0, len(e.PartTags))
		partKeys := make([]string, 0, len(e.PartTags))
		for k := range e.PartTags {
			partKeys = append(partKeys, k)
		}
		slices.Sort(partKeys)
		for _, k := range partKeys {
			partStrs = append(partStrs, k+":"+strings.Join(e.PartTags[k], ","))
		}
		partField := strings.Join(partStrs, ";")
		return line(
			"item",
			field("item", e.Item),
			field("name_key", e.NameKey),
			field("description_key", e.DescriptionKey),
			optionalField("tags", tags),
			optionalField("part_tags", partField),
		)
	default:
		return line("unknown", field("kind", event.EventKind()))
	}
}

func line(eventKind string, fields ...string) string {
	parts := make([]string, 0, len(fields)+1)
	parts = append(parts, field("event", eventKind))
	for _, value := range fields {
		if value != "" {
			parts = append(parts, value)
		}
	}
	return strings.Join(parts, "\t") + "\n"
}

func field(name string, value string) string {
	return name + "=" + escapeValue(value)
}

func optionalField(name string, value string) string {
	if value == "" {
		return ""
	}
	return field(name, value)
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return strings.ReplaceAll(value, "\n", "\\n")
}
