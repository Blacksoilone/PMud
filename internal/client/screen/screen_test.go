package screen_test

import (
	"errors"
	"strings"
	"testing"

	"PMud/internal/client/layout"
	"PMud/internal/client/screen"
)

var errWriteFailed = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

func TestRendererDrawWritesFullRedraw(t *testing.T) {
	var output strings.Builder
	renderer := screen.NewRenderer(&output)
	block := layout.NewBlock([]string{"hello", "world"})

	err := renderer.Draw(block)

	if err != nil {
		t.Fatalf("Draw: %v", err)
	}
	want := "\x1b[2J\x1b[Hhello\nworld\n"
	if output.String() != want {
		t.Fatalf("output = %q, want %q", output.String(), want)
	}
}

func TestRendererDrawPropagatesWriteError(t *testing.T) {
	renderer := screen.NewRenderer(failingWriter{})
	block := layout.NewBlock([]string{"hello"})

	err := renderer.Draw(block)

	if !errors.Is(err, errWriteFailed) {
		t.Fatalf("Draw error = %v, want %v", err, errWriteFailed)
	}
}
