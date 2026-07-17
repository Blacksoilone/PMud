package client

import (
	"errors"
	"io"
	"sync"

	"PMud/internal/client/screen"
	"PMud/internal/client/tui"
	"PMud/internal/protocol"
)

var ErrTUIExit = errors.New("tui exit requested")

type TUIRuntimeConfig struct {
	State        *State
	Output       io.Writer
	Width        int
	Height       int
	HistoryLimit int
}

type TUIRuntime struct {
	state    *State
	model    tui.Model
	renderer screen.Renderer
	width    int
	height   int
	mu       sync.Mutex
}

const defaultTUIHeight = 26

func NewTUIRuntime(config TUIRuntimeConfig) *TUIRuntime {
	return &TUIRuntime{
		state:    config.State,
		model:    tui.NewModel(config.HistoryLimit),
		renderer: screen.NewRenderer(config.Output),
		width:    config.Width,
		height:   normalizeTUIHeight(config.Height),
	}
}

func normalizeTUIHeight(height int) int {
	if height <= 0 {
		return defaultTUIHeight
	}
	return height
}

func (r *TUIRuntime) RenderEvent(event protocol.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.model = tui.ApplyEvent(r.model, event)
	return r.draw()
}

func (r *TUIRuntime) ObserveEvent(event protocol.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state.Observe(event)
	r.model = tui.ApplyEvent(r.model, event)
	return r.draw()
}

func (r *TUIRuntime) SubmitLine(line string, server io.Writer) error {
	if err := r.ApplyInput(tui.Input{Kind: tui.InputText, Text: line}, server); err != nil {
		return err
	}
	return r.ApplyInput(tui.Input{Kind: tui.InputSubmit}, server)
}

func (r *TUIRuntime) ApplyInput(input tui.Input, server io.Writer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var command tui.Command
	r.model, command = tui.ApplyInput(r.model, input)
	if command.ExitRequested {
		return ErrTUIExit
	}
	if command.Submitted {
		resolution := r.state.ResolveCommandInput(command.Line)
		if !resolution.Send {
			r.model = tui.ApplyEvent(r.model, resolution.LocalEvent)
			return r.draw()
		}
		if _, err := io.WriteString(server, resolution.Command+"\n"); err != nil {
			return err
		}
	}
	return r.draw()
}

func (r *TUIRuntime) ForceRedraw() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.draw()
}

func (r *TUIRuntime) Resize(width int, height int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.width = width
	r.height = normalizeTUIHeight(height)
	return r.draw()
}

func (r *TUIRuntime) draw() error {
	return r.renderer.Draw(tui.ViewWithSize(r.model, r.state.catalog, r.width, r.height))
}
