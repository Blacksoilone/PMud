package session

import (
	"PMud/internal/command"
	"PMud/internal/presentation"
	"bufio"
	"io"
	"net"
)

func handleConn(conn net.Conn) error {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		cmd := command.Parse(line)

		event := presentation.InputDebugEvent{
			Verb: cmd.Verb,
			Args: cmd.Args,
		}
		renderer := presentation.TextRenderer{}
		response := renderer.RenderInputDebug(event)
		_, err := io.WriteString(conn, response)
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
