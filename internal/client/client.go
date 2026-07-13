package client

import (
	"PMud/internal/client/render"
	"PMud/internal/content"
	"PMud/internal/protocol"
	"bufio"
	"errors"
	"fmt"
	"io"
)

var ErrProtocolLine = errors.New("protocol line")

func RenderProtocolLines(input io.Reader, output io.Writer, catalog content.ClientCatalog) error {
	return renderProtocolLines(input, output, catalog, nil)
}

func RenderObservedProtocolLines(input io.Reader, output io.Writer, state *State) error {
	return renderProtocolLines(input, output, state.catalog, state)
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

func ForwardCommands(input io.Reader, server io.Writer) error {
	return forwardCommands(input, server, nil)
}

func ForwardResolvedCommands(input io.Reader, server io.Writer, state *State) error {
	return forwardCommands(input, server, state)
}

func forwardCommands(input io.Reader, server io.Writer, state *State) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if state != nil {
			line = state.ResolveCommand(line)
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
