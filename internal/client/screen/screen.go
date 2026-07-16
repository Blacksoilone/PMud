package screen

import (
	"PMud/internal/client/layout"
	"io"
)

const FullRedrawPrefix = "\x1b[2J\x1b[H"

const enterAlternateScreen = "\x1b[?1049h\x1b[?25l"
const exitAlternateScreen = "\x1b[?25h\x1b[?1049l"

type Renderer struct {
	output io.Writer
}

func NewRenderer(output io.Writer) Renderer {
	return Renderer{output: output}
}

func EnterAlternateScreen(output io.Writer) error {
	_, err := io.WriteString(output, enterAlternateScreen)
	return err
}

func ExitAlternateScreen(output io.Writer) error {
	_, err := io.WriteString(output, exitAlternateScreen)
	return err
}

func (r Renderer) Draw(block layout.Block) error {
	if _, err := io.WriteString(r.output, FullRedrawPrefix); err != nil {
		return err
	}
	_, err := io.WriteString(r.output, block.String())
	return err
}
