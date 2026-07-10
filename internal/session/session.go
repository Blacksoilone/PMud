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
		return presentation.SystemMessageEvent{Message: "你没有输入任何内容"}
	}
	if trimmed == "look" {
		return roomObservationEvent(s.game, s.currentRoom)
	}
	if trimmed == "help" {
		return presentation.SystemMessageEvent{Message: "可用命令: look, go <direction>, get <item>, drop <item>, inventory, help\n方向: north/n/北, south/s/南"}
	}
	if remainder, ok := strings.CutPrefix(trimmed, "go "); ok {
		direction := normalizeDirection(strings.TrimSpace(remainder))
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
		return presentation.InventoryEvent{Items: itemIDStrings(s.game.InventoryItemIDs(s.playerID))}
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
		return presentation.SystemMessageEvent{Message: "你迷失在不存在的地方。"}
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
