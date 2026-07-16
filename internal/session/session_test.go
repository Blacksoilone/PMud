package session

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"PMud/internal/presentation"
	"PMud/internal/world"
)

func TestSessionQuest_reportsTutorialStatus(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("quest")

	quest, ok := event.(presentation.QuestStatusEvent)
	if !ok {
		t.Fatalf("expected quest status, got %T", event)
	}
	if quest.QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("quest id = %q", quest.QuestID)
	}
	if quest.QuestName != "教程任务" {
		t.Fatalf("quest name = %q", quest.QuestName)
	}
	if quest.StageID != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage id = %q", quest.StageID)
	}
	if quest.StageText != "拿起旧油灯。" {
		t.Fatalf("stage text = %q", quest.StageText)
	}
	if len(quest.Conditions) != 1 || quest.Conditions[0] != "获取旧油灯" {
		t.Fatalf("conditions = %#v", quest.Conditions)
	}
	if quest.State != "active" {
		t.Fatalf("state = %q", quest.State)
	}
}

func TestSessionActions_advanceTutorialQuest(t *testing.T) {
	state := newTestSessionState()

	state.handleLine("get 旧油灯")
	afterGet := state.handleLine("quest").(presentation.QuestStatusEvent)
	if afterGet.StageID != "quest.tutorial.first_steps.stage.enter_yard" {
		t.Fatalf("stage after get = %q", afterGet.StageID)
	}

	state.handleLine("go north")
	afterMove := state.handleLine("quest").(presentation.QuestStatusEvent)
	if afterMove.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("stage after move = %q", afterMove.StageID)
	}
	if afterMove.State != "active" {
		t.Fatalf("state after move = %q", afterMove.State)
	}

	state.handleLine("examine 练习木剑")
	afterExamine := state.handleLine("quest").(presentation.QuestStatusEvent)
	if afterExamine.StageID != "quest.tutorial.first_steps.stage.examine_sword" {
		t.Fatalf("stage after examine = %q", afterExamine.StageID)
	}
	if afterExamine.State != "reward_pending" {
		t.Fatalf("state after examine = %q", afterExamine.State)
	}
}

func TestSessionHelp_returnsCommandSummary(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("help")

	message, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if message.MessageKey != "system.help" {
		t.Fatalf("expected help message key, got %q", message.MessageKey)
	}
}

func TestSessionUnknownCommand_returnsMessageKeyWithInput(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("dance")

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
			state := newTestSessionState()
			state.currentRoom = tt.fromRoom

			event := state.handleLine(tt.command)

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

func TestNormalizeDirection_mapsStandardAliases(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		want      string
	}{
		{name: "north", direction: "n", want: "north"},
		{name: "south", direction: "s", want: "south"},
		{name: "east", direction: "e", want: "east"},
		{name: "west", direction: "w", want: "west"},
		{name: "up", direction: "u", want: "up"},
		{name: "down", direction: "d", want: "down"},
		{name: "northeast", direction: "ne", want: "northeast"},
		{name: "northwest", direction: "nw", want: "northwest"},
		{name: "southeast", direction: "se", want: "southeast"},
		{name: "southwest", direction: "sw", want: "southwest"},
		{name: "keeps special exits", direction: "trapdoor", want: "trapdoor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDirection(tt.direction)

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSessionGet_resolvesVisibleItemPhrase(t *testing.T) {
	state := newTestSessionState()

	nameEvent := state.handleLine("get 旧油灯")

	nameMessage, ok := nameEvent.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", nameEvent)
	}
	if nameMessage.MessageKey != "system.item.taken" {
		t.Fatalf("expected taken for display name command, got %q", nameMessage.MessageKey)
	}
	if nameMessage.Fields["item"] != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id for display name command, got %q", nameMessage.Fields["item"])
	}

	idState := newTestSessionState()
	idEvent := idState.handleLine("get item.tutorial.old_lantern")
	idMessage, ok := idEvent.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", idEvent)
	}
	if idMessage.MessageKey != "system.item.taken" {
		t.Fatalf("expected taken for item id command, got %q", idMessage.MessageKey)
	}
	if idMessage.Fields["item"] != "item.tutorial.old_lantern" {
		t.Fatalf("expected item id field, got %q", idMessage.Fields["item"])
	}
}

func TestSessionDrop_resolvesInventoryItemPhrase(t *testing.T) {
	state := newTestSessionState()
	state.handleLine("get item.tutorial.old_lantern")

	nameEvent := state.handleLine("drop 旧油灯")

	nameMessage, ok := nameEvent.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", nameEvent)
	}
	if nameMessage.MessageKey != "system.item.dropped" {
		t.Fatalf("expected dropped for display name command, got %q", nameMessage.MessageKey)
	}
	if nameMessage.Fields["item"] != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id for display name command, got %q", nameMessage.Fields["item"])
	}

	idState := newTestSessionState()
	idState.handleLine("get item.tutorial.old_lantern")
	idEvent := idState.handleLine("drop item.tutorial.old_lantern")
	idMessage, ok := idEvent.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", idEvent)
	}
	if idMessage.MessageKey != "system.item.dropped" {
		t.Fatalf("expected dropped for item id command, got %q", idMessage.MessageKey)
	}
	if idMessage.Fields["item"] != "item.tutorial.old_lantern" {
		t.Fatalf("expected item id field, got %q", idMessage.Fields["item"])
	}
}

func TestSessionExamine_resolvesVisibleItemPhrase(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("examine 旧油灯")

	observation, ok := event.(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("expected item observation, got %T", event)
	}
	if observation.Item != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", observation.Item)
	}
	if observation.DescriptionKey != "item.tutorial.old_lantern.description" {
		t.Fatalf("expected old lantern description key, got %q", observation.DescriptionKey)
	}
}

func TestSessionExamine_resolvesAliasPhrase(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("examine jiuyoudeng")

	observation, ok := event.(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("expected item observation, got %T", event)
	}
	if observation.Item != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", observation.Item)
	}
}

func TestSessionExamine_returnsNotHereForInvisibleItem(t *testing.T) {
	state := newTestSessionState()

	event := state.handleLine("examine item.tutorial.practice_sword")

	message, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if message.MessageKey != "system.item.not_here" {
		t.Fatalf("expected not_here message key, got %q", message.MessageKey)
	}
}

func TestHandleConn_writesInitialRoomObservation(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	done := make(chan error, 1)
	go func() {
		done <- handleConn(serverConn, world.New())
	}()

	if err := clientConn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 512)
	n, err := clientConn.Read(buffer)
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

	if _, err := io.WriteString(clientConn, "inventory\n"); err != nil {
		t.Fatal(err)
	}
	n, err := clientConn.Read(buffer)
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
