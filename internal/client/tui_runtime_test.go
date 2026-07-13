package client

import (
	"PMud/internal/client/screen"
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
