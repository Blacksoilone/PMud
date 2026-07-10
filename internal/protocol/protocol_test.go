package protocol

import (
	"errors"
	"testing"
)

func TestParseLine_parsesStructuredEventFields(t *testing.T) {
	// Given
	line := "event=room\troom=room.tutorial.start\tname_key=room.tutorial.start.name\texits=north\titems=item.tutorial.old_lantern\n"

	// When
	event, err := ParseLine(line)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	if event.Name != "room" {
		t.Fatalf("expected event name %q, got %q", "room", event.Name)
	}
	wantFields := map[string]string{
		"room":     "room.tutorial.start",
		"name_key": "room.tutorial.start.name",
		"exits":    "north",
		"items":    "item.tutorial.old_lantern",
	}
	for key, want := range wantFields {
		if got := event.Fields[key]; got != want {
			t.Fatalf("expected field %q to be %q, got %q", key, want, got)
		}
	}
}

func TestParseLine_unescapesFieldValues(t *testing.T) {
	// Given
	line := "event=system\tmessage=第一行\\n第二行\\t反斜杠\\\\\n"

	// When
	event, err := ParseLine(line)

	// Then
	if err != nil {
		t.Fatal(err)
	}
	want := "第一行\n第二行\t反斜杠\\"
	if got := event.Fields["message"]; got != want {
		t.Fatalf("expected message %q, got %q", want, got)
	}
}

func TestParseLine_rejectsMissingEventField(t *testing.T) {
	// Given
	line := "message=hello\n"

	// When
	_, err := ParseLine(line)

	// Then
	if !errors.Is(err, ErrMissingEventField) {
		t.Fatalf("expected ErrMissingEventField, got %v", err)
	}
}

func TestParseLine_rejectsMalformedField(t *testing.T) {
	// Given
	line := "event=system\tmessage\n"

	// When
	_, err := ParseLine(line)

	// Then
	if !errors.Is(err, ErrMalformedField) {
		t.Fatalf("expected ErrMalformedField, got %v", err)
	}
}

func TestParseLine_rejectsUnknownEscape(t *testing.T) {
	// Given
	line := "event=system\tmessage=bad\\xescape\n"

	// When
	_, err := ParseLine(line)

	// Then
	if !errors.Is(err, ErrInvalidEscape) {
		t.Fatalf("expected ErrInvalidEscape, got %v", err)
	}
}

func TestParseLine_rejectsDuplicateFields(t *testing.T) {
	// Given
	line := "event=system\tmessage=hello\tmessage=again\n"

	// When
	_, err := ParseLine(line)

	// Then
	if !errors.Is(err, ErrDuplicateField) {
		t.Fatalf("expected ErrDuplicateField, got %v", err)
	}
}
