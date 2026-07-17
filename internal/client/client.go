package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"PMud/internal/client/keyinput"
	"PMud/internal/client/render"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

var ErrProtocolLine = errors.New("protocol line")

func RenderProtocolLines(input io.Reader, output io.Writer, catalog content.ClientCatalog) error {
	return renderProtocolLines(input, output, catalog, nil)
}

func RenderObservedProtocolLines(input io.Reader, output io.Writer, state *State) error {
	return renderProtocolLines(input, output, state.catalog, state)
}

func RenderTUIProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error {
	return renderTUIProtocolLines(input, output, state, width, historyLimit, false)
}

func RenderTUIObservedProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error {
	return renderTUIProtocolLines(input, output, state, width, historyLimit, true)
}

func RenderTUIObservedProtocolLinesWithRuntime(input io.Reader, runtime *TUIRuntime) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if err := runtime.ObserveEvent(event); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func renderProtocolLines(input io.Reader, output io.Writer, catalog content.ClientCatalog, state *State) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if state != nil {
			state.Observe(event)
		}
		_, err = io.WriteString(output, render.Render(event, catalog))
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func renderTUIProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int, observe bool) error {
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: output, Width: width, HistoryLimit: historyLimit})
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if observe {
			err = runtime.ObserveEvent(event)
		} else {
			err = runtime.RenderEvent(event)
		}
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func ForwardCommands(input io.Reader, server io.Writer) error {
	return forwardCommands(input, server, nil)
}

func ForwardResolvedCommands(input io.Reader, server io.Writer, state *State) error {
	return forwardCommands(input, server, state)
}

func ForwardTUILines(input io.Reader, server io.Writer, runtime *TUIRuntime) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if err := runtime.SubmitLine(scanner.Text(), server); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func ForwardTUIKeyInput(input io.Reader, server io.Writer, runtime *TUIRuntime) error {
	buffer := make([]byte, 256)
	var decoder keyinput.Decoder
	for {
		count, err := input.Read(buffer)
		if count > 0 {
			if err := forwardTUIKeyActions(decoder.Feed(buffer[:count]), server, runtime); err != nil {
				return err
			}
		}
		if err == io.EOF {
			if err := forwardTUIKeyActions(decoder.Flush(), server, runtime); err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func forwardTUIKeyActions(actions []keyinput.Action, server io.Writer, runtime *TUIRuntime) error {
	for _, action := range actions {
		if err := runtime.ApplyInput(action.Input, server); err != nil {
			return err
		}
	}
	return nil
}

func forwardCommands(input io.Reader, server io.Writer, state *State) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if state != nil {
			resolution := state.ResolveCommandInput(line)
			if !resolution.Send {
				continue
			}
			line = resolution.Command
		}
		_, err := io.WriteString(server, line+"\n")
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
