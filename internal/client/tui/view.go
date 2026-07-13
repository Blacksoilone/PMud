package tui

import (
	"PMud/internal/client/layout"
	"PMud/internal/client/render"
	"PMud/internal/content"
)

func View(model Model, catalog content.ClientCatalog, width int) layout.Block {
	return layout.VBox(
		layout.Box(eventHistoryBlock(model, catalog), width),
		layout.Box(promptBlock(model), width),
	)
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
