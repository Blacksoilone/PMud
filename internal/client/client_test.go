package client

import (
	"PMud/internal/content"
	"PMud/internal/protocol"
	"errors"
	"strings"
	"testing"
)

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
	want := "练习场入口\n这里是练习场的入口。北边传来木剑碰撞的声音。\n出口: north\n你看到: 旧油灯\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
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

func TestForwardResolvedCommands_writesResolvedItemIDsToServer(t *testing.T) {
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
	input := strings.NewReader("get 旧油灯\nlook\n")
	var output strings.Builder

	err = ForwardResolvedCommands(input, &output, state)

	if err != nil {
		t.Fatal(err)
	}
	want := "get item.tutorial.old_lantern\nlook\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
