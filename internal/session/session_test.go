package session

import (
	"PMud/internal/presentation"
	"PMud/internal/world"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestSessionHelp_returnsCommandSummary(t *testing.T) {
	// Given
	state := newTestSessionState()

	// When
	event := state.handleLine("help")

	// Then
	message, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if message.MessageKey != "system.help" {
		t.Fatalf("expected help message key, got %q", message.MessageKey)
	}
}

func TestSessionUnknownCommand_returnsMessageKeyWithInput(t *testing.T) {
	// Given
	state := newTestSessionState()

	// When
	event := state.handleLine("dance")

	// Then
	message, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if message.MessageKey != "system.unknown_command" {
		t.Fatalf("expected unknown command message key, got %q", message.MessageKey)
	}
	if message.Fields["input"] != "dance" {
		t.Fatalf("expected input field, got %q", message.Fields["input"])
	}
}

func TestSessionDirectionAliases_moveBetweenRooms(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		fromRoom world.RoomID
		wantRoom string
	}{
		{name: "go n moves north", command: "go n", fromRoom: "room.tutorial.start", wantRoom: "练习场"},
		{name: "go 北 moves north", command: "go 北", fromRoom: "room.tutorial.start", wantRoom: "练习场"},
		{name: "go s moves south", command: "go s", fromRoom: "room.tutorial.yard", wantRoom: "练习场入口"},
		{name: "go 南 moves south", command: "go 南", fromRoom: "room.tutorial.yard", wantRoom: "练习场入口"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			state := newTestSessionState()
			state.currentRoom = tt.fromRoom

			// When
			event := state.handleLine(tt.command)

			// Then
			observation, ok := event.(presentation.RoomObservationEvent)
			if !ok {
				t.Fatalf("expected room observation, got %T", event)
			}
			if observation.Name != tt.wantRoom {
				t.Fatalf("expected room %q, got %q", tt.wantRoom, observation.Name)
			}
		})
	}
}

func TestHandleConn_writesInitialRoomObservation(t *testing.T) {
	// Given
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	done := make(chan error, 1)
	go func() {
		done <- handleConn(serverConn, world.New())
	}()

	// When
	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 512)
	n, err := clientConn.Read(buffer)

	// Then
	if err != nil {
		t.Fatalf("expected initial room output, got %v", err)
	}
	output := string(buffer[:n])
	if !strings.HasPrefix(output, "event=room\t") {
		t.Fatalf("expected initial output to be a room event line, got %q", output)
	}
	if !strings.Contains(output, "\troom=room.tutorial.start\t") {
		t.Fatalf("expected initial output to include start room id field, got %q", output)
	}
	if !strings.Contains(output, "\tname_key=room.tutorial.start.name\t") {
		t.Fatalf("expected initial output to include start room name key field, got %q", output)
	}
	if !strings.Contains(output, "\tdescription_key=room.tutorial.start.description\t") {
		t.Fatalf("expected initial output to include start room description key field, got %q", output)
	}
	if !strings.Contains(output, "\texits=north\t") {
		t.Fatalf("expected initial output to include exits field, got %q", output)
	}
	if !strings.Contains(output, "\titems=item.tutorial.old_lantern\n") {
		t.Fatalf("expected initial output to include items field, got %q", output)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil && err != io.ErrClosedPipe {
		t.Fatalf("expected connection to close cleanly, got %v", err)
	}
}

func TestHandleConn_acceptsExistingCommandsAndWritesStructuredResponses(t *testing.T) {
	// Given
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	done := make(chan error, 1)
	go func() {
		done <- handleConn(serverConn, world.New())
	}()

	if err := clientConn.SetDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 512)
	if _, err := clientConn.Read(buffer); err != nil {
		t.Fatalf("expected initial room output, got %v", err)
	}

	// When
	if _, err := io.WriteString(clientConn, "inventory\n"); err != nil {
		t.Fatal(err)
	}
	n, err := clientConn.Read(buffer)

	// Then
	if err != nil {
		t.Fatalf("expected inventory output, got %v", err)
	}
	output := string(buffer[:n])
	if output != "event=inventory\titems=\n" {
		t.Fatalf("expected structured empty inventory output, got %q", output)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil && err != io.ErrClosedPipe {
		t.Fatalf("expected connection to close cleanly, got %v", err)
	}
}

func newTestSessionState() sessionState {
	game := world.New()
	return sessionState{
		game:        game,
		currentRoom: game.StartRoom(),
		playerID:    "player.local",
	}
}
