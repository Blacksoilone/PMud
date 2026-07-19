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

	event := requireSingleSessionEvent(t, state.handleLine("quest"))

	list, ok := event.(presentation.QuestListEvent)
	if !ok {
		t.Fatalf("expected quest list, got %T", event)
	}
	if len(list.Quests) == 0 {
		t.Fatal("expected at least one quest")
	}
	q := list.Quests[0]
	if q.QuestID != "quest.tutorial.first_steps" {
		t.Fatalf("quest id = %q", q.QuestID)
	}
	if q.QuestName != "教程任务" {
		t.Fatalf("quest name = %q", q.QuestName)
	}
	if q.StageID != "quest.tutorial.first_steps.stage.get_lantern" {
		t.Fatalf("stage id = %q", q.StageID)
	}
	if q.StageText != "拿起旧油灯。" {
		t.Fatalf("stage text = %q", q.StageText)
	}
	if len(q.Conditions) != 1 || q.Conditions[0] != "获取旧油灯" {
		t.Fatalf("conditions = %#v", q.Conditions)
	}
	if q.State != "active" {
		t.Fatalf("state = %q", q.State)
	}
}

func TestSessionActions_advanceTutorialQuest(t *testing.T) {
	state := newTestSessionState()

	state.handleLine("go east")      // hall → lock_hall
	state.handleLine("get 旧油灯")     // get_lantern → enter_chamber
	afterGetList := requireSingleSessionEvent(t, state.handleLine("quest")).(presentation.QuestListEvent)
	if len(afterGetList.Quests) == 0 {
		t.Fatal("expected quests in list")
	}
	if afterGetList.Quests[0].StageID != "quest.tutorial.first_steps.stage.enter_chamber" {
		t.Fatalf("stage after get = %q", afterGetList.Quests[0].StageID)
	}

	state.handleLine("go east")      // lock_hall → lock_chamber (trigger moved_room → examine_relic)
	afterMoveList := requireSingleSessionEvent(t, state.handleLine("quest")).(presentation.QuestListEvent)
	if len(afterMoveList.Quests) == 0 {
		t.Fatal("expected quests in list")
	}
	if afterMoveList.Quests[0].StageID != "quest.tutorial.first_steps.stage.examine_relic" {
		t.Fatalf("stage after move = %q", afterMoveList.Quests[0].StageID)
	}
	if afterMoveList.Quests[0].State != "active" {
		t.Fatalf("state after move = %q", afterMoveList.Quests[0].State)
	}

	state.handleLine("examine 练功徽章")
	afterExamineList := requireSingleSessionEvent(t, state.handleLine("quest")).(presentation.QuestListEvent)
	if len(afterExamineList.Quests) == 0 {
		t.Fatal("expected quests in list")
	}
	if afterExamineList.Quests[0].StageID != "quest.tutorial.first_steps.stage.examine_relic" {
		t.Fatalf("stage after examine = %q", afterExamineList.Quests[0].StageID)
	}
	if afterExamineList.Quests[0].State != "completed" {
		t.Fatalf("state after examine = %q", afterExamineList.Quests[0].State)
	}
}

func TestSessionActions_returnQuestProgressNotificationAfterAdvancingQuest(t *testing.T) {
	state := newTestSessionState()

	state.handleLine("go east")
	events := state.handleLine("get 旧油灯")

	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}
	taken, ok := events[0].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("first event type = %T, want presentation.SystemMessageEvent", events[0])
	}
	if taken.MessageKey != "system.item.taken" {
		t.Fatalf("first message key = %q, want system.item.taken", taken.MessageKey)
	}
	assertQuestProgressEvent(t, events[1], "quest.tutorial.first_steps.stage.enter_chamber", "active")
	if _, ok := events[2].(presentation.QuestStatusEvent); !ok {
		t.Fatalf("third event type = %T, want presentation.QuestStatusEvent", events[2])
	}
}

func TestSessionActions_returnQuestCompletedNotificationAfterFinalStage(t *testing.T) {
	state := newTestSessionState()
	state.handleLine("go east")      // hall → lock_hall
	state.handleLine("get 旧油灯")     // get_lantern → enter_chamber
	state.handleLine("go east")      // lock_hall → lock_chamber, enter_chamber → examine_relic

	events := state.handleLine("examine 练功徽章")

	if len(events) != 3 {
		t.Fatalf("event count = %d, want 3", len(events))
	}
	item, ok := events[0].(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("first event type = %T, want presentation.ItemObservationEvent", events[0])
	}
	if item.Item != "item.tutorial.training_relic" {
		t.Fatalf("item = %q, want item.tutorial.training_relic", item.Item)
	}
	assertQuestProgressEvent(t, events[1], "quest.tutorial.first_steps.stage.examine_relic", "completed")
	if _, ok := events[2].(presentation.QuestStatusEvent); !ok {
		t.Fatalf("third event type = %T, want presentation.QuestStatusEvent", events[2])
	}
}

func TestSessionActions_doNotReturnQuestProgressNotificationWhenQuestDoesNotAdvance(t *testing.T) {
	state := newTestSessionState()

	state.handleLine("go east")   // hall → lock_hall
	state.handleLine("get 旧油灯") // get_lantern → enter_chamber (needs moved_room to lock_chamber)
	state.handleLine("go west")   // back to hall
	state.handleLine("go north")  // hall → item_yard (does NOT trigger enter_chamber)

	events := state.handleLine("go south")
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	if _, ok := events[0].(presentation.RoomObservationEvent); !ok {
		t.Fatalf("event type = %T, want presentation.RoomObservationEvent", events[0])
	}
}

func TestSessionLockedDoor_returnsSystemMoveLocked(t *testing.T) {
	state := newTestSessionState()

	state.handleLine("go east")
	event := requireSingleSessionEvent(t, state.handleLine("go east"))

	msg, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if msg.MessageKey != "system.move.locked" {
		t.Fatalf("message key = %q, want system.move.locked", msg.MessageKey)
	}
}

func TestSessionActions_lookReturnsRoomObservationAndLogFeedback(t *testing.T) {
	state := newTestSessionState()

	events := state.handleLine("look")

	if len(events) != 2 {
		t.Fatalf("event count = %d, want 2", len(events))
	}
	if _, ok := events[0].(presentation.RoomObservationEvent); !ok {
		t.Fatalf("first event type = %T, want presentation.RoomObservationEvent", events[0])
	}
	feedback, ok := events[1].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("second event type = %T, want presentation.SystemMessageEvent", events[1])
	}
	if feedback.MessageKey != "system.look.observed" {
		t.Fatalf("feedback message key = %q, want system.look.observed", feedback.MessageKey)
	}
}

func TestSessionHelp_returnsCommandSummary(t *testing.T) {
	state := newTestSessionState()

	event := requireSingleSessionEvent(t, state.handleLine("help"))

	message, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", event)
	}
	if message.MessageKey != "system.help" {
		t.Fatalf("expected help message key, got %q", message.MessageKey)
	}
}

func TestSessionHandleLine_returnsSingleEventListForHelp(t *testing.T) {
	state := newTestSessionState()

	events := state.handleLine("help")

	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	message, ok := events[0].(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("expected system message, got %T", events[0])
	}
	if message.MessageKey != "system.help" {
		t.Fatalf("expected help message key, got %q", message.MessageKey)
	}
}

func TestSessionUnknownCommand_returnsMessageKeyWithInput(t *testing.T) {
	state := newTestSessionState()

	event := requireSingleSessionEvent(t, state.handleLine("dance"))

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
		setup    []string // commands to run before checking fromRoom
		wantRoom string
	}{
		{name: "go n moves north", command: "go n", setup: nil, fromRoom: "room.tutorial.hall", wantRoom: "物品庭院"},
		{name: "go 北 moves north", command: "go 北", setup: nil, fromRoom: "room.tutorial.hall", wantRoom: "物品庭院"},
		{name: "go s moves south", command: "go s", setup: []string{"go north"}, fromRoom: "room.tutorial.item_yard", wantRoom: "教学大厅"},
		{name: "go 南 moves south", command: "go 南", setup: []string{"go north"}, fromRoom: "room.tutorial.item_yard", wantRoom: "教学大厅"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := newTestSessionState()
			for _, cmd := range tt.setup {
				state.handleLine(cmd)
			}
			state.currentRoom = tt.fromRoom

			event := requireFirstSessionEvent(t, state.handleLine(tt.command))

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
	state.handleLine("go east")

	nameEvent := requireFirstSessionEvent(t, state.handleLine("get 旧油灯"))

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
	idState.handleLine("go east")
	idEvent := requireFirstSessionEvent(t, idState.handleLine("get item.tutorial.old_lantern"))
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
	state.handleLine("go east")
	state.handleLine("get item.tutorial.old_lantern")

	nameEvent := requireSingleSessionEvent(t, state.handleLine("drop 旧油灯"))

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
	idState.handleLine("go east")
	idState.handleLine("get item.tutorial.old_lantern")
	idEvent := requireSingleSessionEvent(t, idState.handleLine("drop item.tutorial.old_lantern"))
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
	state.handleLine("go east")

	event := requireSingleSessionEvent(t, state.handleLine("examine 旧油灯"))

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
	state.handleLine("go east")

	event := requireSingleSessionEvent(t, state.handleLine("examine jiuyoudeng"))

	observation, ok := event.(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("expected item observation, got %T", event)
	}
	if observation.Item != "item.tutorial.old_lantern" {
		t.Fatalf("expected old lantern id, got %q", observation.Item)
	}
}

func TestSessionLook_resolvesItemPhrase(t *testing.T) {
	state := newTestSessionState()
	state.handleLine("go east")
	state.handleLine("get 旧油灯")
	state.handleLine("go west")
	state.handleLine("go north")

	event := requireFirstSessionEvent(t, state.handleLine("look practice-sword"))

	observation, ok := event.(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("expected item observation, got %T", event)
	}
	if observation.Item != "item.tutorial.practice_sword" {
		t.Fatalf("expected practice sword id, got %q", observation.Item)
	}
}

func TestSessionExamine_resolvesPracticeSwordPinyinAlias(t *testing.T) {
	state := newTestSessionState()
	state.handleLine("go east")
	state.handleLine("get 旧油灯")
	state.handleLine("go west")
	state.handleLine("go north")

	event := requireFirstSessionEvent(t, state.handleLine("examine lianximujian"))

	observation, ok := event.(presentation.ItemObservationEvent)
	if !ok {
		t.Fatalf("expected item observation, got %T", event)
	}
	if observation.Item != "item.tutorial.practice_sword" {
		t.Fatalf("expected practice sword id, got %q", observation.Item)
	}
}

func TestSessionExamine_returnsNotHereForInvisibleItem(t *testing.T) {
	state := newTestSessionState()

	event := requireSingleSessionEvent(t, state.handleLine("examine item.tutorial.practice_sword"))

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
	loop := newTestLoop()
	go func() {
		done <- handleConn(serverConn, loop)
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
	if !strings.Contains(output, "\troom=room.tutorial.hall\t") {
		t.Fatalf("expected initial output to include hall room id field, got %q", output)
	}
	if !strings.Contains(output, "\tname_key=room.tutorial.hall.name\t") {
		t.Fatalf("expected initial output to include hall name key field, got %q", output)
	}
	if !strings.Contains(output, "\tdescription_key=room.tutorial.hall.description\t") {
		t.Fatalf("expected initial output to include hall description key field, got %q", output)
	}
	if !strings.Contains(output, "\texits=") {
		t.Fatalf("expected initial output to include exits field, got %q", output)
	}
	if !strings.Contains(output, "\titems=\n") {
		t.Fatalf("expected initial output to include empty items field, got %q", output)
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
	loop := newTestLoop()
	go func() {
		done <- handleConn(serverConn, loop)
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

func newTestLoop() *world.Loop {
	game := world.New()
	loop := world.NewLoop(game)
	loop.Start()
	return loop
}

func newTestSessionState() sessionState {
	game := world.New()
	loop := world.NewLoop(game)
	loop.Start()
	loop.EnterWorld("player.local")

	outgoing := make(chan []presentation.Event, 8)
	loop.Register("player.local", outgoing)

	return sessionState{
		loop:        loop,
		incoming:    outgoing,
		currentRoom: game.StartRoom(),
		playerID:    "player.local",
	}
}

func requireSingleSessionEvent(t *testing.T, events []presentation.Event) presentation.Event {
	t.Helper()
	if len(events) != 1 {
		t.Fatalf("event count = %d, want 1", len(events))
	}
	return events[0]
}

func requireFirstSessionEvent(t *testing.T, events []presentation.Event) presentation.Event {
	t.Helper()
	if len(events) == 0 {
		t.Fatal("event count = 0, want at least 1")
	}
	return events[0]
}

func assertQuestProgressEvent(t *testing.T, event presentation.Event, wantStageID, wantState string) {
	t.Helper()
	progress, ok := event.(presentation.SystemMessageEvent)
	if !ok {
		t.Fatalf("progress event type = %T, want presentation.SystemMessageEvent", event)
	}
	if progress.MessageKey != "system.quest.progress" {
		t.Fatalf("progress message key = %q, want system.quest.progress", progress.MessageKey)
	}
	if progress.Fields["quest_id"] != "quest.tutorial.first_steps" {
		t.Fatalf("quest_id = %q, want quest.tutorial.first_steps", progress.Fields["quest_id"])
	}
	if progress.Fields["stage_id"] != wantStageID {
		t.Fatalf("stage_id = %q, want %q", progress.Fields["stage_id"], wantStageID)
	}
	if progress.Fields["state"] != wantState {
		t.Fatalf("state = %q, want %q", progress.Fields["state"], wantState)
	}
}
