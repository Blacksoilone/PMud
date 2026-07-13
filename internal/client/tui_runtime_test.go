package client

import (
	"PMud/internal/client/screen"
	"PMud/internal/client/tui"
	"PMud/internal/content"
	"PMud/internal/protocol"
	"strings"
	"testing"
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
	if serverOutput.String() != "get item.tutorial.old_lantern\n" {
		t.Fatalf("server output = %q", serverOutput.String())
	}
	got := screenOutput.String()
	if strings.Count(got, screen.FullRedrawPrefix) != 2 {
		t.Fatalf("redraw count = %d, want 2; output:\n%s", strings.Count(got, screen.FullRedrawPrefix), got)
	}
	if !strings.Contains(got, "> get 旧油灯") || !strings.Contains(got, "| >") {
		t.Fatalf("output missing submitted or cleared prompt:\n%s", got)
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
	if !strings.Contains(lastFrame, "| >") {
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
	if server.String() != "get item.tutorial.old_lantern\n" {
		t.Fatalf("server output = %q", server.String())
	}
	got := output.String()
	lastFrame := got[strings.LastIndex(got, screen.FullRedrawPrefix):]
	if !strings.Contains(got, "> get 旧油灯") {
		t.Fatalf("output missing typed prompt:\n%s", got)
	}
	if !strings.Contains(lastFrame, "| >") {
		t.Fatalf("last frame missing cleared prompt:\n%s", lastFrame)
	}
}
