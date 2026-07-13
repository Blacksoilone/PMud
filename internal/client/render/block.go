package render

import (
	"PMud/internal/client/layout"
	"PMud/internal/content"
	"PMud/internal/protocol"
	"strings"
)

func RenderBlock(event protocol.Event, catalog content.ClientCatalog) layout.Block {
	return layout.NewBlock(strings.Split(strings.TrimSuffix(Render(event, catalog), "\n"), "\n"))
}
