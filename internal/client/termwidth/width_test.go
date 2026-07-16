package termwidth

import "testing"

func TestWidth_countsASCIIAsOneCell(t *testing.T) {
	got := Width("hello")
	want := 5
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestWidth_countsCJKAsTwoCells(t *testing.T) {
	got := Width("旧油灯")
	want := 6
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestWidth_countsMixedText(t *testing.T) {
	got := Width("HP:旧油灯")
	want := 9
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestWidth_ignoresANSIEscapeSequences(t *testing.T) {
	got := Width("\x1b[48;5;240m旧油灯\x1b[0m")
	want := 6
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestLineWidth_returnsLongestRenderedLine(t *testing.T) {
	got := LineWidth("abc\n旧油灯")
	want := 6
	if got != want {
		t.Fatalf("expected %d, got %d", want, got)
	}
}

func TestRightPad_padsToTargetCellWidth(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		targetWidth int
		want        string
	}{
		{name: "pads ASCII text", input: "HP", targetWidth: 5, want: "HP   "},
		{name: "pads CJK text by rendered cells", input: "旧油灯", targetWidth: 8, want: "旧油灯  "},
		{name: "pads mixed ASCII and CJK text", input: "HP:旧油灯", targetWidth: 12, want: "HP:旧油灯   "},
		{name: "pads ANSI styled text by visible width", input: "\x1b[48;5;240m旧油灯\x1b[0m", targetWidth: 8, want: "\x1b[48;5;240m旧油灯\x1b[0m  "},
		{name: "leaves exact width unchanged", input: "旧油灯", targetWidth: 6, want: "旧油灯"},
		{name: "leaves already overwide text unchanged", input: "旧油灯", targetWidth: 4, want: "旧油灯"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := RightPad(test.input, test.targetWidth)
			if got != test.want {
				t.Fatalf("expected %q, got %q", test.want, got)
			}
		})
	}
}
