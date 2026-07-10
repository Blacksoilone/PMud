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
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
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
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		_, err := io.WriteString(server, scanner.Text()+"\n")
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
