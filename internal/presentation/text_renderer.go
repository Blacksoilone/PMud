package presentation

import "strings"

type TextRenderer struct{}

func (TextRenderer) Render(event Event) string {
	switch e := event.(type) {
	case SystemMessageEvent:
		return e.Message + "\n"
	case RoomObservationEvent:
		var builder strings.Builder
		builder.WriteString(e.Name)
		builder.WriteString("\n")
		builder.WriteString(e.Description)
		builder.WriteString("\n")
		if len(e.Exits) > 0 {
			builder.WriteString("出口: ")
			builder.WriteString(strings.Join(e.Exits, ", "))
			builder.WriteString("\n")
		}
		if len(e.Items) > 0 {
			builder.WriteString("你看到: ")
			builder.WriteString(strings.Join(e.Items, ", "))
			builder.WriteString("\n")
		}
		return builder.String()
	case InventoryEvent:
		if len(e.Items) == 0 {
			return "你什么也没有带。\n"
		}
		var builder strings.Builder
		builder.WriteString("你带着: ")
		builder.WriteString(strings.Join(e.Items, ", "))
		builder.WriteString("\n")
		return builder.String()
	default:
		return "未知事件类型: " + e.EventKind() + "\n"
	}
}
