package screen

import (
	"PMud/internal/client/layout"
	"io"
)

const FullRedrawPrefix = "\x1b[2J\x1b[H"

type Renderer struct {
	output io.Writer
}

func NewRenderer(output io.Writer) Renderer {
	return Renderer{output: output}
}

func (r Renderer) Draw(block layout.Block) error {
	if _, err := io.WriteString(r.output, FullRedrawPrefix); err != nil {
		return err
	}
	_, err := io.WriteString(r.output, block.String())
	return err
}
