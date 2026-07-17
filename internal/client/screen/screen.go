package screen

import (
	"io"
	"strings"

	"PMud/internal/client/layout"
)

const (
	FullRedrawPrefix      = "\x1b[2J\x1b[H"
	OverwriteRedrawPrefix = "\x1b[H"
)

const (
	enterAlternateScreen = "\x1b[?1049h\x1b[?25l"
	exitAlternateScreen  = "\x1b[?25h\x1b[?1049l"
)

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
	if _, err := io.WriteString(r.output, OverwriteRedrawPrefix); err != nil {
		return err
	}
	frame := strings.TrimSuffix(block.String(), "\n")
	_, err := io.WriteString(r.output, strings.ReplaceAll(frame, "\n", "\r\n"))
	return err
}
