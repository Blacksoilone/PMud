package tui

import (
	"PMud/internal/client/layout"
	"PMud/internal/client/render"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

const rightHUDWidth = 36

const rightHUDMinWidth = 96

func View(model Model, catalog content.ClientCatalog, width int) layout.Block {
	if width < rightHUDMinWidth {
		return legacyView(model, catalog, width)
	}
	mainWidth := width - rightHUDWidth
	mainColumn := layout.VBox(
		paneBlock("房间 / 可见物", roomPaneBlock(model, catalog)),
		paneBlock("日志", eventHistoryBlock(model, catalog)),
	)
	rightHUD := layout.VBox(
		paneBlock("小地图", minimapPaneBlock(model, catalog)),
		paneBlock("状态", statusPaneBlock()),
		paneBlock("当前任务", questPaneBlock(model)),
	)
	return layout.VBox(
		layout.HBox(0, layout.Box(mainColumn, mainWidth), layout.Box(rightHUD, rightHUDWidth)),
		layout.Box(promptBlock(model), width),
	)
}

func legacyView(model Model, catalog content.ClientCatalog, width int) layout.Block {
	return layout.VBox(
		layout.Box(eventHistoryBlock(model, catalog), width),
		layout.Box(promptBlock(model), width),
	)
}

func paneBlock(title string, body layout.Block) layout.Block {
	lines := make([]string, 0, len(body.Lines)+1)
	lines = append(lines, title)
	lines = append(lines, body.Lines...)
	return layout.NewBlock(lines)
}

func roomPaneBlock(model Model, catalog content.ClientCatalog) layout.Block {
	room := model.Regions.Room
	if room.Room == "" {
		return layout.NewBlock([]string{"尚未观察房间"})
	}
	return render.RenderBlock(protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            room.Room,
			"name_key":        room.NameKey,
			"description_key": room.DescriptionKey,
			"exits":           room.Exits,
			"items":           room.Items,
		},
	}, catalog)
}

func minimapPaneBlock(model Model, catalog content.ClientCatalog) layout.Block {
	region := MinimapRegion{
		AreaName: "当前区域",
		Current:  MinimapRoom{Label: minimapLabel(model, catalog)},
	}
	lines := []string{region.AreaName}
	lines = append(lines, renderMinimapGrid(region)...)
	return layout.NewBlock(lines)
}

func minimapLabel(model Model, catalog content.ClientCatalog) string {
	nameKey := model.Regions.Room.NameKey
	if nameKey == "" {
		return "当前"
	}
	name, ok := catalog.Text[content.TextKey(nameKey)]
	if !ok || name == "" {
		return "当前"
	}
	return name
}

func statusPaneBlock() layout.Block {
	return layout.NewBlock([]string{"状态系统未开放"})
}

func questPaneBlock(model Model) layout.Block {
	quest := model.Regions.Quest
	if quest.QuestID == "" {
		return layout.NewBlock([]string{"未追踪任务"})
	}
	return layout.NewBlock([]string{
		quest.QuestName,
		"阶段: " + quest.StageText,
		"目标: " + quest.Conditions,
	})
}

func eventHistoryBlock(model Model, catalog content.ClientCatalog) layout.Block {
	blocks := make([]layout.Block, 0, len(model.Events))
	for _, event := range model.Events {
		blocks = append(blocks, render.RenderBlock(event, catalog))
	}
	return layout.VBox(blocks...)
}

func promptBlock(model Model) layout.Block {
	return layout.NewBlock([]string{"> " + model.Input})
}
