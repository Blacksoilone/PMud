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
		return presentation.SystemMessageEvent{Message: "你没有输入任何内容"}
	}
	if trimmed == "look" {
		return roomObservationEvent(s.game, s.currentRoom)
	}
	if remainder, ok := strings.CutPrefix(trimmed, "go "); ok {
		direction := strings.TrimSpace(remainder)
		nextRoom, ok := s.game.Move(s.currentRoom, direction)
		if !ok {
			return presentation.SystemMessageEvent{Message: "你不能往那个方向走。"}
		}
		s.currentRoom = nextRoom
		return roomObservationEvent(s.game, s.currentRoom)
	}
	if remainder, ok := strings.CutPrefix(trimmed, "get "); ok {
		itemName := strings.TrimSpace(remainder)
		_, ok := s.game.GetItem(s.currentRoom, itemName, s.playerID)
		if !ok {
			return presentation.SystemMessageEvent{Message: "这里没有那个东西。"}
		}
		return presentation.SystemMessageEvent{Message: "你拿起了" + itemName + "。"}
	}
	if trimmed == "inventory" {
		return presentation.InventoryEvent{Items: s.game.Inventory(s.playerID)}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "drop "); ok {
		itemName := strings.TrimSpace(remainder)
		ok := s.game.DropItemByName(s.currentRoom, itemName, s.playerID)
		if !ok {
			return presentation.SystemMessageEvent{Message: "你没有那个东西。"}
		}
		return presentation.SystemMessageEvent{Message: "你放下了" + itemName + "。"}
	}

	return presentation.SystemMessageEvent{Message: "你输入了: " + line}
}

func roomObservationEvent(game *world.World, roomID world.RoomID) presentation.Event {
	observation, ok := game.Look(roomID)
	if !ok {
		return presentation.SystemMessageEvent{Message: "你迷失在不存在的地方。"}
	}
	return presentation.RoomObservationEvent{
		Name:        observation.Name,
		Description: observation.Description,
		Exits:       observation.Exits,
		Items:       observation.Items,
	}
}
