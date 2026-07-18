package session

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"

	"PMud/internal/command"
	"PMud/internal/presentation"
	"PMud/internal/world"
)

var playerIDCounter atomic.Uint64

func handleConn(conn net.Conn, loop *world.Loop) error {
	defer conn.Close()

	playerID := world.PlayerID(fmt.Sprintf("player.%d", playerIDCounter.Add(1)))
	_, ok := loop.EnterWorld(playerID)
	if !ok {
		return errors.New("failed to create player in world")
	}
	defer loop.LeaveWorld(playerID)

	outgoing := make(chan []presentation.Event, 8)
	loop.Register(playerID, outgoing)
	defer loop.Unregister(playerID)

	renderer := presentation.TextRenderer{}

	resp := make(chan world.ActionResult, 1)
	loop.Submit(world.Action{
		PlayerID: playerID,
		Verb:     "look",
		Resp:     resp,
	})
	initResult := <-resp
	state := sessionState{
		loop:        loop,
		incoming:    outgoing,
		currentRoom: initResult.NewRoom,
		playerID:    playerID,
	}
	writeEvents(conn, &renderer, initResult.Events)

	type lineOrDone struct {
		line string
	}
	lines := make(chan lineOrDone, 8)
	scannerDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			lines <- lineOrDone{line: scanner.Text()}
		}
		scannerDone <- scanner.Err()
	}()

	for {
		select {
		case ld := <-lines:
			if err := writeEvents(conn, &renderer, state.handleLine(ld.line)); err != nil {
				return err
			}
		case err := <-scannerDone:
			return err
		case broadcast := <-state.incoming:
			if err := writeEvents(conn, &renderer, broadcast); err != nil {
				return err
			}
		}
	}
}

func writeEvents(conn net.Conn, renderer *presentation.TextRenderer, events []presentation.Event) error {
	var buf strings.Builder
	for _, event := range events {
		buf.WriteString(renderer.Render(event))
	}
	_, err := io.WriteString(conn, buf.String())
	return err
}

type sessionState struct {
	loop        *world.Loop
	incoming    chan []presentation.Event
	currentRoom world.RoomID
	playerID    world.PlayerID
}

func (s *sessionState) handleLine(line string) []presentation.Event {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.empty_input"})
	}

	parsed := command.ParseServerInput(line)
	switch cmd := parsed.(type) {
	case command.LookCommand:
		return s.submitAction("look", "")
	case command.HelpCommand:
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.help"})
	case command.QuestCommand:
		return s.submitAction("quest", "")
	case command.MoveCommand:
		dir := normalizeDirection(cmd.Direction)
		return s.submitAction("move", dir)
	case command.InventoryCommand:
		return s.submitAction("inventory", "")
	case command.ItemCommand:
		return s.handleItemCommand(cmd)
	case command.UnknownCommand:
		return singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.unknown_command",
			Fields:     map[string]string{"input": cmd.Input},
		})
	default:
		return singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.unknown_command",
			Fields:     map[string]string{"input": line},
		})
	}
}

func (s *sessionState) submitAction(verb, target string) []presentation.Event {
	resp := make(chan world.ActionResult, 1)
	s.loop.Submit(world.Action{
		PlayerID: s.playerID,
		Verb:     verb,
		Target:   target,
		Resp:     resp,
	})
	result := <-resp
	s.currentRoom = result.NewRoom
	return result.Events
}

func (s *sessionState) handleItemCommand(itemCommand command.ItemCommand) []presentation.Event {
	switch itemCommand.Verb {
	case command.ItemVerbGet:
		return s.submitAction("get", itemCommand.Target)
	case command.ItemVerbDrop:
		return s.submitAction("drop", itemCommand.Target)
	case command.ItemVerbExamine:
		return s.submitAction("examine", itemCommand.Target)
	case command.ItemVerbLook:
		return s.submitAction("look-item", itemCommand.Target)
	default:
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.unknown_command"})
	}
}

func singleEvent(event presentation.Event) []presentation.Event {
	return []presentation.Event{event}
}

func normalizeDirection(direction string) string {
	canonical, ok := command.CanonicalDirection(direction)
	if ok {
		return canonical
	}
	return direction
}
