package client

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"PMud/internal/client/screen"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

type signaledWriter struct {
	mu     sync.Mutex
	output strings.Builder
	writes chan struct{}
}

func (w *signaledWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	w.output.Write(data)
	w.mu.Unlock()
	if string(data) == screen.OverwriteRedrawPrefix {
		select {
		case w.writes <- struct{}{}:
		default:
		}
	}
	return len(data), nil
}

func (w *signaledWriter) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.output.String()
}

type gatedReader struct {
	chunks       [][]byte
	index        int
	firstRead    chan struct{}
	continueRead chan struct{}
	release      chan struct{}
	once         sync.Once
}

func (r *gatedReader) Read(data []byte) (int, error) {
	if r.index < len(r.chunks) {
		chunk := r.chunks[r.index]
		r.index++
		if r.index == 1 && r.firstRead != nil {
			r.once.Do(func() { close(r.firstRead) })
			<-r.continueRead
		}
		return copy(data, chunk), nil
	}
	<-r.release
	return 0, io.EOF
}

func TestRenderProtocolLines_rendersServerEvents(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = RenderProtocolLines(input, &output, compiled.Client)
	if err != nil {
		t.Fatal(err)
	}
	want := "练习场入口\n这里是练习场的入口。北边传来木剑碰撞的声音。\n出口: north\n你看到: 旧油灯（old lantern）\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRenderProtocolLines_rendersQuestProgressAfterCommandResult(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("event=system\tmessage_key=system.item.taken\titem=item.tutorial.old_lantern\n" +
		"event=system\tmessage_key=system.quest.progress\tquest_id=quest.tutorial.first_steps\tstage_id=quest.tutorial.first_steps.stage.enter_yard\tstate=active\n")
	var output strings.Builder

	err = RenderProtocolLines(input, &output, compiled.Client)
	if err != nil {
		t.Fatal(err)
	}
	got := output.String()
	if !strings.Contains(got, "你拿起了旧油灯（old lantern）。") {
		t.Fatalf("output missing item result:\n%s", got)
	}
	if !strings.Contains(got, "任务更新: active") {
		t.Fatalf("output missing quest progress:\n%s", got)
	}
}

func TestRenderProtocolLines_returnsParseError(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("event=system\tmessage\n")
	var output strings.Builder

	err = RenderProtocolLines(input, &output, compiled.Client)

	if !errors.Is(err, ErrProtocolLine) {
		t.Fatalf("expected ErrProtocolLine, got %v", err)
	}
}

func TestRenderTUIProtocolLines_rendersObservedEvents(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = RenderTUIProtocolLines(input, &output, state, 128, 3)
	if err != nil {
		t.Fatalf("RenderTUIProtocolLines: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, "练习场入口") {
		t.Fatalf("output does not include room name:\n%s", got)
	}
	if !strings.Contains(got, "> ") {
		t.Fatalf("output does not include prompt:\n%s", got)
	}
}

func TestRenderTUIObservedProtocolLines_updatesCommandResolution(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = RenderTUIObservedProtocolLines(input, &output, state, 128, 3)
	if err != nil {
		t.Fatalf("RenderTUIObservedProtocolLines: %v", err)
	}
	line := state.ResolveCommand("get 旧油灯")
	if line != "get 旧油灯" {
		t.Fatalf("resolved command = %q", line)
	}
}

func TestRenderTUIObservedProtocolLines_redrawsPerEvent(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("event=system\tmessage_key=system.help\n" +
		"event=inventory\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	err = RenderTUIObservedProtocolLines(input, &output, state, 128, 3)
	if err != nil {
		t.Fatalf("RenderTUIObservedProtocolLines: %v", err)
	}
	got := output.String()
	redrawCount := strings.Count(got, screen.OverwriteRedrawPrefix)
	if redrawCount != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", redrawCount, got)
	}
	if !strings.Contains(got, "可用命令") || !strings.Contains(got, "你带着: 旧油灯（old lantern）") || !strings.Contains(got, "> ") {
		t.Fatalf("output missing expected TUI content:\n%s", got)
	}
}

func TestForwardCommands_writesInputLinesToServer(t *testing.T) {
	input := strings.NewReader("look\ngo n\n")
	var output strings.Builder

	err := ForwardCommands(input, &output)
	if err != nil {
		t.Fatal(err)
	}
	want := "look\ngo n\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestForwardResolvedCommands_forwardsUnresolvedItemPhrasesToServer(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("get 旧油灯\nlook\n")
	var output strings.Builder

	err = ForwardResolvedCommands(input, &output, state)
	if err != nil {
		t.Fatal(err)
	}
	want := "get 旧油灯\nlook\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestForwardResolvedCommands_forwardsAmbiguousItemPhraseToServer(t *testing.T) {
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("get shared\nlook\n")
	var output strings.Builder

	err = ForwardResolvedCommands(input, &output, state)
	if err != nil {
		t.Fatal(err)
	}
	want := "get shared\nlook\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestForwardTUILines_redrawsInputAndWritesResolvedCommand(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"items": "item.tutorial.old_lantern",
		},
	})
	input := strings.NewReader("get 旧油灯\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = ForwardTUILines(input, &serverOutput, runtime)
	if err != nil {
		t.Fatalf("ForwardTUILines: %v", err)
	}
	if serverOutput.String() != "get 旧油灯\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if !strings.Contains(got, "> get 旧油灯") || !strings.Contains(got, "> ") {
		t.Fatalf("screen output missing submitted or cleared prompt:\n%s", got)
	}
}

func TestRenderTUIObservedProtocolLinesWithRuntime_sharesModelWithForwardTUILines(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	serverEvents := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	input := strings.NewReader("get 旧油灯\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = RenderTUIObservedProtocolLinesWithRuntime(serverEvents, runtime)
	if err != nil {
		t.Fatalf("RenderTUIObservedProtocolLinesWithRuntime: %v", err)
	}
	err = ForwardTUILines(input, &serverOutput, runtime)
	if err != nil {
		t.Fatalf("ForwardTUILines: %v", err)
	}
	if serverOutput.String() != "get 旧油灯\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if !strings.Contains(got, "练习场入口") || !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("screen output missing shared event history or input:\n%s", got)
	}
}

func TestForwardTUIKeyInput_decodesTextBackspaceAndSubmit(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{Name: "room", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})
	input := strings.NewReader("get 旧油灯\x7f灯\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = ForwardTUIKeyInput(input, &serverOutput, runtime)
	if err != nil {
		t.Fatalf("ForwardTUIKeyInput: %v", err)
	}
	if serverOutput.String() != "get 旧油灯\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("screen output missing typed or cleared prompt:\n%s", got)
	}
}

func TestForwardTUIKeyInput_decodesClear(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("look\x15inventory\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = ForwardTUIKeyInput(input, &serverOutput, runtime)
	if err != nil {
		t.Fatalf("ForwardTUIKeyInput: %v", err)
	}
	if serverOutput.String() != "inventory\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	if !strings.Contains(screenOutput.String(), "> ") {
		t.Fatalf("screen output missing cleared prompt:\n%s", screenOutput.String())
	}
}

func TestForwardTUIKeyInput_ignoresCtrlCAndContinuesInput(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("look\x03inventory\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = ForwardTUIKeyInput(input, &serverOutput, runtime)
	if err != nil {
		t.Fatalf("ForwardTUIKeyInput: %v", err)
	}
	if serverOutput.String() != "lookinventory\n" {
		t.Fatalf("server output = %q, want lookinventory command", serverOutput.String())
	}
	if !strings.Contains(screenOutput.String(), "lookinventory") {
		t.Fatalf("screen output missing input after ignored Ctrl+C:\n%s", screenOutput.String())
	}
}

func TestForwardTUIKeyInput_confirmedQuitReturnsExitSignal(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := strings.NewReader("quit\ny\n")
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 128, HistoryLimit: 3})

	err = ForwardTUIKeyInput(input, &serverOutput, runtime)

	if !errors.Is(err, ErrTUIExit) {
		t.Fatalf("error = %v, want ErrTUIExit", err)
	}
	if serverOutput.String() != "" {
		t.Fatalf("server output = %q, want empty", serverOutput.String())
	}
}

func TestForwardTUIKeyInput_closesPopupAfterLoneEscapeTimeout(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := &gatedReader{
		chunks:  [][]byte{{'?'}, {0x1b}},
		release: make(chan struct{}),
	}
	output := &signaledWriter{writes: make(chan struct{}, 4)}
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: output, Width: 128, HistoryLimit: 3})

	done := make(chan error, 1)
	go func() { done <- ForwardTUIKeyInput(input, &serverOutput, runtime) }()

	for range 2 {
		select {
		case <-output.writes:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for popup open and escape timeout redraws")
		}
	}
	close(input.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ForwardTUIKeyInput: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("ForwardTUIKeyInput did not finish")
	}
	got := output.String()
	if strings.Count(got, "帮助") == 0 {
		t.Fatalf("output never rendered help popup:\n%s", got)
	}
	if strings.Contains(got[strings.LastIndex(got, screen.OverwriteRedrawPrefix):], "帮助") {
		t.Fatalf("lone escape did not close popup in latest frame:\n%s", got)
	}
}

func TestForwardTUIKeyInput_keepsSplitFunctionKeyTogetherWithinEscapeTimeout(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	input := &gatedReader{
		chunks:       [][]byte{{0x1b}, []byte("[11~")},
		firstRead:    make(chan struct{}),
		continueRead: make(chan struct{}),
		release:      make(chan struct{}),
	}
	output := &signaledWriter{writes: make(chan struct{}, 4)}
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: output, Width: 128, HistoryLimit: 3})

	done := make(chan error, 1)
	go func() { done <- ForwardTUIKeyInput(input, &serverOutput, runtime) }()
	select {
	case <-input.firstRead:
		close(input.continueRead)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for split escape prefix")
	}
	select {
	case <-output.writes:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for F1 redraw")
	}
	close(input.release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ForwardTUIKeyInput: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("ForwardTUIKeyInput did not finish")
	}
	if got := output.String(); !strings.Contains(got, "帮助") {
		t.Fatalf("split F1 sequence did not open help popup:\n%s", got)
	}
}
