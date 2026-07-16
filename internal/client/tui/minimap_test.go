package tui

import (
	"strings"
	"testing"

	"PMud/internal/client/termwidth"
)

func TestRenderMinimapGrid_centersLabelsInFixedSlots(t *testing.T) {
	region := MinimapRegion{
		AreaName: "演示-校场",
		Current:  MinimapRoom{Label: "练功场"},
		Neighbors: map[MapDirection]MinimapRoom{
			MapNorthwest: {Label: "民兵营房"},
			MapNorth:     {Label: "武器库房"},
			MapNortheast: {Label: "精锐营房"},
			MapWest:      {Label: "井"},
			MapEast:      {Label: "将军府"},
			MapSouthwest: {Label: "塔楼a"},
			MapSouth:     {Label: "大门"},
			MapSoutheast: {Label: "塔楼b"},
		},
	}

	lines := renderMinimapGrid(region)

	want := []string{
		"民兵营房  武器库房  精锐营房",
		"         \\   ｜   /         ",
		"   井   -- 练功场 -- 将军府 ",
		"         /   ｜   \\         ",
		" 塔楼a      大门     塔楼b  ",
	}
	if len(lines) != len(want) {
		t.Fatalf("line count = %d, want %d: %#v", len(lines), len(want), lines)
	}
	for index := range want {
		plain := termwidth.StripANSI(lines[index])
		if plain != want[index] {
			t.Fatalf("line %d = %q, want %q", index, plain, want[index])
		}
		if termwidth.Width(lines[index]) != minimapGridWidth {
			t.Fatalf("line %d width = %d, want %d", index, termwidth.Width(lines[index]), minimapGridWidth)
		}
	}
	assertMinimapCellsHaveGrayBackground(t, lines[0], 3)
	assertMinimapCellsHaveGrayBackground(t, lines[2], 3)
	assertMinimapCellsHaveGrayBackground(t, lines[4], 3)
}

func TestRenderMinimapGrid_omitsMissingNeighborsAndConnectors(t *testing.T) {
	region := MinimapRegion{
		AreaName: "教程-练习场",
		Current:  MinimapRoom{Label: "入口"},
		Neighbors: map[MapDirection]MinimapRoom{
			MapNorth:     {Label: "练习院"},
			MapNortheast: {Label: "器械房"},
		},
	}

	lines := renderMinimapGrid(region)

	want := []string{
		"           练习院    器械房 ",
		"             ｜   /         ",
		"            入口            ",
		"                            ",
		"                            ",
	}
	for index := range want {
		plain := termwidth.StripANSI(lines[index])
		if plain != want[index] {
			t.Fatalf("line %d = %q, want %q", index, plain, want[index])
		}
		if termwidth.Width(lines[index]) != minimapGridWidth {
			t.Fatalf("line %d width = %d, want %d", index, termwidth.Width(lines[index]), minimapGridWidth)
		}
	}
	assertMinimapCellsHaveGrayBackground(t, lines[0], 2)
	assertMinimapCellsHaveGrayBackground(t, lines[2], 1)
	assertMinimapCellsHaveGrayBackground(t, lines[4], 0)
}

func TestRenderMinimapGrid_truncatesLongLabelsToFixedCell(t *testing.T) {
	region := MinimapRegion{
		Current: MinimapRoom{Label: "练习场入口"},
		Neighbors: map[MapDirection]MinimapRoom{
			MapNorth: {Label: "很长很长的房间名"},
		},
	}

	lines := renderMinimapGrid(region)

	for index, line := range lines {
		if termwidth.Width(line) != minimapGridWidth {
			t.Fatalf("line %d width = %d, want %d: %q", index, termwidth.Width(line), minimapGridWidth, line)
		}
	}
	if termwidth.StripANSI(lines[0]) != "          很长很长          " {
		t.Fatalf("top row = %q, want long label truncated into center cell", lines[0])
	}
	if termwidth.StripANSI(lines[2]) != "          练习场入          " {
		t.Fatalf("middle row = %q, want current label truncated into center cell", lines[2])
	}
}

func assertMinimapCellsHaveGrayBackground(t *testing.T, line string, wantCount int) {
	t.Helper()
	gotCount := strings.Count(line, minimapCellBackground)
	if gotCount != wantCount {
		t.Fatalf("gray background count = %d, want %d in %q", gotCount, wantCount, line)
	}
}
