package session

import (
	"bufio"
	"io"
	"net"
	"strings"

	"PMud/internal/command"
	"PMud/internal/presentation"
	"PMud/internal/progression"
	"PMud/internal/world"
)

func handleConn(conn net.Conn, game *world.World) error {
	defer conn.Close()
	renderer := presentation.TextRenderer{} // 复用renderer
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
		events := state.handleLine(scanner.Text())
		for _, event := range events {
			response := renderer.Render(event)
			_, err := io.WriteString(conn, response)
			if err != nil {
				return err
			}
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
	progression *progression.Engine
}

func (s *sessionState) handleLine(line string) []presentation.Event {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.empty_input"})
	}

	parsed := command.ParseServerInput(line)
	switch command := parsed.(type) {
	case command.LookCommand:
		return singleEvent(roomObservationEvent(s.game, s.currentRoom))
	case command.HelpCommand:
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.help"})
	case command.QuestCommand:
		return singleEvent(s.questStatusEvent())
	case command.MoveCommand:
		nextRoom, ok := s.game.Move(s.currentRoom, command.Direction)
		if !ok {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.move.blocked"})
		}
		s.currentRoom = nextRoom
		events := singleEvent(roomObservationEvent(s.game, s.currentRoom))
		events = append(events, s.applyProgression(progression.Trigger{Kind: progression.TriggerMovedRoom, RoomID: string(nextRoom)})...)
		return events
	case command.InventoryCommand:
		return singleEvent(presentation.InventoryEvent{Items: itemIDStrings(s.game.InventoryItemIDs(s.playerID))})
	case command.ItemCommand:
		return s.handleItemCommand(command)
	case command.UnknownCommand:
		return singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.unknown_command",
			Fields: map[string]string{
				"input": command.Input,
			},
		})
	default:
		return singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.unknown_command",
			Fields: map[string]string{
				"input": line,
			},
		})
	}
}

func (s *sessionState) handleItemCommand(itemCommand command.ItemCommand) []presentation.Event {
	switch itemCommand.Verb {
	case command.ItemVerbGet:
		resolution := s.game.ResolveRoomItemPhrase(s.currentRoom, itemCommand.Target)
		if len(resolution.AmbiguousItemIDs) > 0 {
			return singleEvent(ambiguousItemEvent(s.game, resolution.AmbiguousItemIDs))
		}
		if !resolution.Found {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		itemID, ok := s.game.GetItem(s.currentRoom, resolution.ItemID, s.playerID)
		if !ok {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		events := singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.item.taken",
			Fields: map[string]string{
				"item": string(itemID),
			},
		})
		events = append(events, s.applyProgression(progression.Trigger{Kind: progression.TriggerGotItem, ItemID: string(itemID)})...)
		return events
	case command.ItemVerbDrop:
		resolution := s.game.ResolveInventoryItemPhrase(s.playerID, itemCommand.Target)
		if len(resolution.AmbiguousItemIDs) > 0 {
			return singleEvent(ambiguousItemEvent(s.game, resolution.AmbiguousItemIDs))
		}
		if !resolution.Found {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"})
		}
		itemID, ok := s.game.DropInventoryItem(s.currentRoom, resolution.ItemID, s.playerID)
		if !ok {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_carried"})
		}
		return singleEvent(presentation.SystemMessageEvent{
			MessageKey: "system.item.dropped",
			Fields: map[string]string{
				"item": string(itemID),
			},
		})
	case command.ItemVerbExamine:
		resolution := s.game.ResolveVisibleItemPhrase(s.currentRoom, s.playerID, itemCommand.Target)
		if len(resolution.AmbiguousItemIDs) > 0 {
			return singleEvent(ambiguousItemEvent(s.game, resolution.AmbiguousItemIDs))
		}
		if !resolution.Found {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		item, ok := s.game.ExamineItem(s.currentRoom, resolution.ItemID, s.playerID)
		if !ok {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		events := singleEvent(itemObservationEvent(item))
		events = append(events, s.applyProgression(progression.Trigger{Kind: progression.TriggerExaminedItem, ItemID: string(item.Item)})...)
		return events
	case command.ItemVerbLook:
		resolution := s.game.ResolveVisibleItemPhrase(s.currentRoom, s.playerID, itemCommand.Target)
		if len(resolution.AmbiguousItemIDs) > 0 {
			return singleEvent(ambiguousItemEvent(s.game, resolution.AmbiguousItemIDs))
		}
		if !resolution.Found {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		item, ok := s.game.ExamineItem(s.currentRoom, resolution.ItemID, s.playerID)
		if !ok {
			return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.item.not_here"})
		}
		return singleEvent(itemObservationEvent(item))
	default:
		return singleEvent(presentation.SystemMessageEvent{MessageKey: "system.unknown_command"})
	}
}

func singleEvent(event presentation.Event) []presentation.Event {
	return []presentation.Event{event}
}

func (s *sessionState) progressionEngine() *progression.Engine {
	if s.progression == nil {
		s.progression = progression.NewEngine(s.game.ProgressionDefinitions())
	}
	return s.progression
}

func (s *sessionState) applyProgression(trigger progression.Trigger) []presentation.Event {
	status, advanced := s.progressionEngine().Apply(string(s.playerID), trigger)
	if !advanced {
		return nil
	}
	if status.State == progression.QuestStateRewardPending {
		resolvedStatus, resolved := s.progressionEngine().ResolveRewards(string(s.playerID))
		if resolved {
			status = resolvedStatus
		}
	}
	return singleEvent(questProgressEvent(status))
}

func questProgressEvent(status progression.Status) presentation.Event {
	return presentation.SystemMessageEvent{
		MessageKey: "system.quest.progress",
		Fields: map[string]string{
			"quest_id": status.QuestID,
			"stage_id": status.StageID,
			"state":    string(status.State),
		},
	}
}

func (s *sessionState) questStatusEvent() presentation.Event {
	status, ok := s.progressionEngine().Status(string(s.playerID))
	if !ok {
		return presentation.SystemMessageEvent{MessageKey: "system.quest.none"}
	}
	return presentation.QuestStatusEvent{
		QuestID:    status.QuestID,
		QuestName:  status.QuestName,
		StageID:    status.StageID,
		StageText:  status.StageText,
		Conditions: status.Conditions,
		State:      string(status.State),
	}
}

func ambiguousItemEvent(game *world.World, itemIDs []world.ItemID) presentation.Event {
	return presentation.SystemMessageEvent{Message: "名字不明确: " + strings.Join(game.ItemNames(itemIDs), ", ")}
}

func itemObservationEvent(item world.ItemObservation) presentation.Event {
	return presentation.ItemObservationEvent{
		Item:           string(item.Item),
		NameKey:        item.NameKey,
		DescriptionKey: item.DescriptionKey,
		Name:           item.Name,
		Description:    item.Description,
	}
}

func normalizeDirection(direction string) string {
	canonical, ok := command.CanonicalDirection(direction)
	if ok {
		return canonical
	}
	return direction
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
