package session

import (
	"PMud/internal/presentation"
	"PMud/internal/world"
	"bufio"
	"io"
	"net"
	"strings"
)

func handleConn(conn net.Conn, game *world.World) error {
	defer conn.Close()
	renderer := presentation.TextRenderer{} //复用renderer
	state := sessionState{
		game:        game,
		currentRoom: game.StartRoom(),
		playerID:    "player.local",
	}
	initialResponse := renderer.Render(roomObservationEvent(state.game, state.currentRoom))
	_, err := io.WriteString(conn, initialResponse)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		event := state.handleLine(scanner.Text())
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

type sessionState struct {
	game        *world.World
	currentRoom world.RoomID
	playerID    world.PlayerID
}

func (s *sessionState) handleLine(line string) presentation.Event {
	trimmed := strings.TrimSpace(line)

	if trimmed == "" {
		return presentation.SystemMessageEvent{MessageKey: "system.empty_input"}
	}
	if trimmed == "look" {
		return roomObservationEvent(s.game, s.currentRoom)
	}
	if trimmed == "help" {
		return presentation.SystemMessageEvent{MessageKey: "system.help"}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "go "); ok {
		direction := normalizeDirection(strings.TrimSpace(remainder))
		nextRoom, ok := s.game.Move(s.currentRoom, direction)
		if !ok {
			return presentation.SystemMessageEvent{MessageKey: "system.move.blocked"}
		}
		s.currentRoom = nextRoom
		return roomObservationEvent(s.game, s.currentRoom)
	}
	if remainder, ok := strings.CutPrefix(trimmed, "get "); ok {
		itemID, ok := s.game.GetItem(s.currentRoom, world.ItemID(strings.TrimSpace(remainder)), s.playerID)
		if !ok {
			return presentation.SystemMessageEvent{MessageKey: "system.item.not_here"}
		}
		return presentation.SystemMessageEvent{
			MessageKey: "system.item.taken",
			Fields: map[string]string{
				"item": string(itemID),
			},
		}
	}
	if trimmed == "inventory" {
		return presentation.InventoryEvent{Items: itemIDStrings(s.game.InventoryItemIDs(s.playerID))}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "drop "); ok {
		itemID, ok := s.game.DropInventoryItem(s.currentRoom, world.ItemID(strings.TrimSpace(remainder)), s.playerID)
		if !ok {
			return presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"}
		}
		return presentation.SystemMessageEvent{
			MessageKey: "system.item.dropped",
			Fields: map[string]string{
				"item": string(itemID),
			},
		}
	}

	return presentation.SystemMessageEvent{
		MessageKey: "system.unknown_command",
		Fields: map[string]string{
			"input": line,
		},
	}
}

func normalizeDirection(direction string) string {
	switch direction {
	case "n", "北":
		return "north"
	case "s", "南":
		return "south"
	default:
		return direction
	}
}

func roomObservationEvent(game *world.World, roomID world.RoomID) presentation.Event {
	observation, ok := game.Look(roomID)
	if !ok {
		return presentation.SystemMessageEvent{MessageKey: "system.room.missing"}
	}
	return presentation.RoomObservationEvent{
		Room:           string(observation.Room),
		NameKey:        observation.NameKey,
		DescriptionKey: observation.DescriptionKey,
		Name:           observation.Name,
		Description:    observation.Description,
		Exits:          observation.Exits,
		Items:          itemIDStrings(observation.ItemIDs),
	}
}

func itemIDStrings(itemIDs []world.ItemID) []string {
	items := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, string(itemID))
	}
	return items
}
