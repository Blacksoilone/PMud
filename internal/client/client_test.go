package client

import (
	"PMud/internal/content"
	"errors"
	"strings"
	"testing"
)

func TestRenderProtocolLines_rendersServerEvents(t *testing.T) {
	// Given
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\tdescription_key=room.tutorial.start.description\texits=north\titems=item.tutorial.old_lantern\n")
	var output strings.Builder

	// When
	err = RenderProtocolLines(input, &output, compiled.Client)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	want := "练习场入口\n这里是练习场的入口。北边传来木剑碰撞的声音。\n出口: north\n你看到: 旧油灯\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestRenderProtocolLines_returnsParseError(t *testing.T) {
	// Given
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		t.Fatal(err)
	}
	input := strings.NewReader("event=system\tmessage\n")
	var output strings.Builder

	// When
	err = RenderProtocolLines(input, &output, compiled.Client)

	// Then
	if !errors.Is(err, ErrProtocolLine) {
		t.Fatalf("expected ErrProtocolLine, got %v", err)
	}
}

func TestForwardCommands_writesInputLinesToServer(t *testing.T) {
	// Given
	input := strings.NewReader("look\ngo n\n")
	var output strings.Builder

	// When
	err := ForwardCommands(input, &output)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	want := "look\ngo n\n"
	if got := output.String(); got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
