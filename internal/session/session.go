package session

import (
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
		event := presentation.EchoEvent{Line: line}
		renderer := presentation.TextRenderer{}
		response := renderer.Render(event)
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
