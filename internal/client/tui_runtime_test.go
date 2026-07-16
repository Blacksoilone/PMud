package client

import (
	"strings"
	"testing"

	"PMud/internal/client/screen"
	"PMud/internal/client/tui"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

func TestTUIRuntimeObserveEventRedrawsScreen(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ObserveEvent(protocol.Event{Name: "inventory", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})
	if err != nil {
		t.Fatalf("ObserveEvent: %v", err)
	}
	got := output.String()
	if !strings.Contains(got, screen.FullRedrawPrefix) || !strings.Contains(got, "你带着: 旧油灯") {
		t.Fatalf("output missing redraw or event text:\n%s", got)
	}
}

func TestTUIRuntimeSubmitLineRedrawsInputThenClearsPrompt(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{Name: "room", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 48, HistoryLimit: 3})

	err = runtime.SubmitLine("get 旧油灯", &serverOutput)
	if err != nil {
		t.Fatalf("SubmitLine: %v", err)
	}
	if serverOutput.String() != "get 旧油灯\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "> get 旧油灯") || !strings.Contains(got, "> ") {
		t.Fatalf("output missing submitted or cleared prompt:\n%s", got)
	}
}

func TestTUIRuntimeSubmitLineWritesCaseInsensitiveAliases(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var screenOutput strings.Builder
	var serverOutput strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &screenOutput, Width: 48, HistoryLimit: 3})

	if err := runtime.SubmitLine("TAKE jiuyoudeng", &serverOutput); err != nil {
		t.Fatalf("SubmitLine TAKE: %v", err)
	}
	if err := runtime.SubmitLine("NW", &serverOutput); err != nil {
		t.Fatalf("SubmitLine NW: %v", err)
	}

	want := "get jiuyoudeng\ngo northwest\n"
	if got := serverOutput.String(); got != want {
		t.Fatalf("server output = %q, want %q", got, want)
	}
}

func TestTUIRuntimeApplyInputRedrawsText(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "get "}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "旧油灯"}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	got := output.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("output missing typed prompt:\n%s", got)
	}
	if server.String() != "" {
		t.Fatalf("server output = %q, want empty", server.String())
	}
}

func TestTUIRuntimeApplyInputBackspaceRemovesLastRune(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "get 旧油灯"}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	err = runtime.ApplyInput(tui.Input{Kind: tui.InputBackspace}, &server)
	if err != nil {
		t.Fatalf("ApplyInput backspace: %v", err)
	}
	got := output.String()
	lastFrame := got[strings.LastIndex(got, screen.FullRedrawPrefix):]
	if !strings.Contains(lastFrame, "> get 旧油") {
		t.Fatalf("last frame missing shortened prompt:\n%s", lastFrame)
	}
	if strings.Contains(lastFrame, "> get 旧油灯") {
		t.Fatalf("last frame still contains removed rune:\n%s", lastFrame)
	}
}

func TestTUIRuntimeApplyInputClearClearsPrompt(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "inventory"}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	err = runtime.ApplyInput(tui.Input{Kind: tui.InputClear}, &server)
	if err != nil {
		t.Fatalf("ApplyInput clear: %v", err)
	}
	got := output.String()
	lastFrame := got[strings.LastIndex(got, screen.FullRedrawPrefix):]
	if !strings.Contains(lastFrame, "> ") {
		t.Fatalf("last frame missing cleared prompt:\n%s", lastFrame)
	}
	if strings.Contains(lastFrame, "inventory") {
		t.Fatalf("last frame still contains cleared input:\n%s", lastFrame)
	}
}

func TestTUIRuntimeApplyInputSubmitWritesResolvedCommandAndClearsPrompt(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	state.Observe(protocol.Event{Name: "room", Fields: map[string]string{"items": "item.tutorial.old_lantern"}})
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "get 旧油灯"}, &server)
	if err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	err = runtime.ApplyInput(tui.Input{Kind: tui.InputSubmit}, &server)
	if err != nil {
		t.Fatalf("ApplyInput submit: %v", err)
	}
	if server.String() != "get 旧油灯\n" {
		t.Fatalf("server output = %q", server.String())
	}
	got := output.String()
	lastFrame := got[strings.LastIndex(got, screen.FullRedrawPrefix):]
	if !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("output missing typed prompt:\n%s", got)
	}
	if !strings.Contains(lastFrame, "> ") {
		t.Fatalf("last frame missing cleared prompt:\n%s", lastFrame)
	}
}

func TestTUIRuntimeSubmitLineForwardsAmbiguousItemPhrase(t *testing.T) {
	compiled, err := content.Compile(ambiguousAliasContentSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.SubmitLine("get shared", &server)
	if err != nil {
		t.Fatalf("SubmitLine: %v", err)
	}
	if got := server.String(); got != "get shared\n" {
		t.Fatalf("server output = %q, want get shared", got)
	}
	if got := output.String(); strings.Contains(got, "名字不明确") {
		t.Fatalf("output should not include local ambiguity message:\n%s", got)
	}
}

func TestTUIRuntimeSubmitLineShowsHelpLocally(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.SubmitLine("help", &server)
	if err != nil {
		t.Fatalf("SubmitLine: %v", err)
	}
	if got := server.String(); got != "" {
		t.Fatalf("server output = %q, want empty", got)
	}
	if got := output.String(); !strings.Contains(got, "可用命令") {
		t.Fatalf("output missing help text:\n%s", got)
	}
	if got := output.String(); !strings.Contains(got, "get/take <item>") || !strings.Contains(got, "examine/x/inspect <item>") {
		t.Fatalf("output missing item command aliases:\n%s", got)
	}
	if got := output.String(); !strings.Contains(got, "northeast/ne") || !strings.Contains(got, "northwest/nw") || !strings.Contains(got, "southeast/se") || !strings.Contains(got, "southwest/sw") {
		t.Fatalf("output missing diagonal direction aliases:\n%s", got)
	}
}

func TestTUIRuntimeSubmitLineShowsEmptyInputLocally(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 48, HistoryLimit: 3})

	err = runtime.SubmitLine("", &server)
	if err != nil {
		t.Fatalf("SubmitLine: %v", err)
	}
	if got := server.String(); got != "" {
		t.Fatalf("server output = %q, want empty", got)
	}
	if got := output.String(); !strings.Contains(got, "你没有输入任何内容") {
		t.Fatalf("output missing empty-input text:\n%s", got)
	}
}

func TestTUIRuntimeForceRedrawDoesNotWriteServerOrChangeInput(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	var server strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 128, Height: 32, HistoryLimit: 3})

	if err := runtime.ApplyInput(tui.Input{Kind: tui.InputText, Text: "look"}, &server); err != nil {
		t.Fatalf("ApplyInput text: %v", err)
	}
	if err := runtime.ApplyInput(tui.Input{Kind: tui.InputForceRedraw}, &server); err != nil {
		t.Fatalf("ApplyInput force redraw: %v", err)
	}

	if server.String() != "" {
		t.Fatalf("server output = %q, want empty", server.String())
	}
	if strings.Count(output.String(), screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2", strings.Count(output.String(), screen.FullRedrawPrefix))
	}
	lastFrame := output.String()[strings.LastIndex(output.String(), screen.FullRedrawPrefix):]
	if !strings.Contains(lastFrame, "> look") {
		t.Fatalf("last frame changed input:\n%s", lastFrame)
	}
}

func TestTUIRuntimeResizeUsesNewHeight(t *testing.T) {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(compiled.Client)
	var output strings.Builder
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: &output, Width: 128, Height: 26, HistoryLimit: 3})

	if err := runtime.Resize(128, 32); err != nil {
		t.Fatalf("Resize: %v", err)
	}

	lastFrame := output.String()[strings.LastIndex(output.String(), screen.FullRedrawPrefix)+len(screen.FullRedrawPrefix):]
	lines := strings.Split(strings.TrimSuffix(lastFrame, "\n"), "\n")
	if len(lines) != 32 {
		t.Fatalf("redrawn line count = %d, want 32", len(lines))
	}
}
