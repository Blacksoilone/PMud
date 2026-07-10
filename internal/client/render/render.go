package render

import (
	"PMud/internal/content"
	"PMud/internal/protocol"
	"strings"
)

func Render(event protocol.Event, catalog content.ClientCatalog) string {
	switch event.Name {
	case "room":
		return renderRoom(event, catalog)
	case "inventory":
		return renderInventory(event, catalog)
	case "system":
		return renderSystem(event, catalog)
	default:
		return "未知事件: " + event.Name + "\n"
	}
}

func renderRoom(event protocol.Event, catalog content.ClientCatalog) string {
	var builder strings.Builder
	builder.WriteString(text(catalog, event.Fields["name_key"]))
	builder.WriteString("\n")
	builder.WriteString(text(catalog, event.Fields["description_key"]))
	builder.WriteString("\n")

	exits := event.Fields["exits"]
	if exits != "" {
		builder.WriteString("出口: ")
		builder.WriteString(strings.ReplaceAll(exits, ",", ", "))
		builder.WriteString("\n")
	}

	items := itemNames(catalog, event.Fields["items"])
	if len(items) > 0 {
		builder.WriteString("你看到: ")
		builder.WriteString(strings.Join(items, ", "))
		builder.WriteString("\n")
	}
	return builder.String()
}

func renderInventory(event protocol.Event, catalog content.ClientCatalog) string {
	items := itemNames(catalog, event.Fields["items"])
	if len(items) == 0 {
		return "你什么也没有带。\n"
	}
	return "你带着: " + strings.Join(items, ", ") + "\n"
}

func renderSystem(event protocol.Event, catalog content.ClientCatalog) string {
	if messageKey := event.Fields["message_key"]; messageKey != "" {
		return text(catalog, messageKey) + "\n"
	}
	return event.Fields["message"] + "\n"
}

func itemNames(catalog content.ClientCatalog, itemIDs string) []string {
	if itemIDs == "" {
		return nil
	}
	ids := strings.Split(itemIDs, ",")
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		itemID := content.ItemID(id)
		nameKey, ok := catalog.ItemNames[itemID]
		if !ok {
			names = append(names, id)
			continue
		}
		names = append(names, text(catalog, string(nameKey)))
	}
	return names
}

func text(catalog content.ClientCatalog, key string) string {
	value, ok := catalog.Text[content.TextKey(key)]
	if !ok {
		return key
	}
	return value
}
