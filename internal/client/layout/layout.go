package layout

import "PMud/internal/client/panel"

type Block struct {
	Lines []string
}

func NewBlock(lines []string) Block {
	return Block{Lines: append([]string(nil), lines...)}
}

func (b Block) String() string {
	return panel.RenderLines(b.Lines)
}

func Box(block Block, width int) Block {
	return Block{Lines: panel.BoxLines(block.Lines, width)}
}

func VBox(blocks ...Block) Block {
	lines := make([]string, 0)
	for _, block := range blocks {
		lines = append(lines, block.Lines...)
	}
	return Block{Lines: lines}
}

func HBox(gap int, blocks ...Block) Block {
	if len(blocks) == 0 {
		return Block{}
	}
	joined := blocks[0].Lines
	for _, block := range blocks[1:] {
		joined = panel.JoinHorizontal(joined, block.Lines, gap)
	}
	return Block{Lines: joined}
}
