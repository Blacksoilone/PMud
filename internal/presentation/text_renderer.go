package presentation

import (
	"slices"
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
		return line(
			"room",
			field("room", e.Room),
			field("name_key", e.NameKey),
			field("description_key", e.DescriptionKey),
			field("exits", strings.Join(e.Exits, ",")),
			field("items", strings.Join(e.Items, ",")),
		)
	case InventoryEvent:
		return line("inventory", field("items", strings.Join(e.Items, ",")))
	default:
		return line("unknown", field("kind", event.EventKind()))
	}
}

func line(eventKind string, fields ...string) string {
	parts := make([]string, 0, len(fields)+1)
	parts = append(parts, field("event", eventKind))
	parts = append(parts, fields...)
	return strings.Join(parts, "\t") + "\n"
}

func field(name string, value string) string {
	return name + "=" + escapeValue(value)
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\t", "\\t")
	return strings.ReplaceAll(value, "\n", "\\n")
}
