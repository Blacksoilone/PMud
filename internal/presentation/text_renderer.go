package presentation

import "strings"

type TextRenderer struct{}

func (TextRenderer) Render(event Event) string {
	switch e := event.(type) {
	case SystemMessageEvent:
		return line("system", field("message", e.Message))
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
