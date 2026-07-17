package tui

import (
	"strings"

	"PMud/internal/client/termwidth"
)

const (
	minimapCellWidth      = 8
	minimapGapWidth       = 2
	minimapGridWidth      = minimapCellWidth*3 + minimapGapWidth*2
	minimapCellGray       = "\x1b[48;5;240m"
	minimapStyleReset     = "\x1b[0m"
	minimapCellBackground = minimapCellGray
)

type MapDirection string

const (
	MapNorthwest MapDirection = "northwest"
	MapNorth     MapDirection = "north"
	MapNortheast MapDirection = "northeast"
	MapWest      MapDirection = "west"
	MapEast      MapDirection = "east"
	MapSouthwest MapDirection = "southwest"
	MapSouth     MapDirection = "south"
	MapSoutheast MapDirection = "southeast"
)

type MinimapRegion struct {
	AreaName  string
	Current   MinimapRoom
	Neighbors map[MapDirection]MinimapRoom
}

type MinimapRoom struct {
	Label string
}

func renderMinimapGrid(region MinimapRegion) []string {
	return []string{
		gridRow(
			region.Neighbors[MapNorthwest].Label,
			region.Neighbors[MapNorth].Label,
			region.Neighbors[MapNortheast].Label,
		),
		connectorRow(region, upperConnectors),
		middleRow(region),
		connectorRow(region, lowerConnectors),
		gridRow(
			region.Neighbors[MapSouthwest].Label,
			region.Neighbors[MapSouth].Label,
			region.Neighbors[MapSoutheast].Label,
		),
	}
}

func gridRow(left string, center string, right string) string {
	return cell(left) + gap("") + cell(center) + gap("") + cell(right)
}

func middleRow(region MinimapRegion) string {
	westGap := ""
	if region.Neighbors[MapWest].Label != "" {
		westGap = "──"
	}
	eastGap := ""
	if region.Neighbors[MapEast].Label != "" {
		eastGap = "──"
	}
	return cell(region.Neighbors[MapWest].Label) + gap(westGap) + cell(region.Current.Label) + gap(eastGap) + cell(region.Neighbors[MapEast].Label)
}

func cell(label string) string {
	label = normalizeMinimapLabel(label)
	width := termwidth.Width(label)
	if label == "" {
		return strings.Repeat(" ", minimapCellWidth)
	}
	var padded string
	if width >= minimapCellWidth {
		padded = termwidth.RightPad(label, minimapCellWidth)
		return minimapCellGray + padded + minimapStyleReset
	}
	left := (minimapCellWidth - width) / 2
	right := minimapCellWidth - width - left
	padded = strings.Repeat(" ", left) + label + strings.Repeat(" ", right)
	return minimapCellGray + padded + minimapStyleReset
}

func normalizeMinimapLabel(label string) string {
	label = truncateMinimapLabel(label)
	if termwidth.Width(label)%2 == 1 {
		return label + " "
	}
	return label
}

func truncateMinimapLabel(label string) string {
	width := 0
	var builder strings.Builder
	for _, char := range label {
		charWidth := termwidth.Width(string(char))
		if width+charWidth > minimapCellWidth {
			break
		}
		builder.WriteRune(char)
		width += charWidth
	}
	return builder.String()
}

func gap(connector string) string {
	if connector == "" {
		return strings.Repeat(" ", minimapGapWidth)
	}
	return termwidth.RightPad(connector, minimapGapWidth)
}

type connectorSet struct {
	left  string
	right string
}

var (
	upperConnectors = connectorSet{left: "└──┐", right: "┌──┘"}
	lowerConnectors = connectorSet{left: "┌──┘", right: "└──┐"}
)

func connectorRow(region MinimapRegion, connectors connectorSet) string {
	left := false
	right := false
	if connectors == upperConnectors {
		if region.Neighbors[MapNorthwest].Label != "" {
			left = true
		}
		if region.Neighbors[MapNortheast].Label != "" {
			right = true
		}
	} else {
		if region.Neighbors[MapSouthwest].Label != "" {
			left = true
		}
		if region.Neighbors[MapSoutheast].Label != "" {
			right = true
		}
	}
	leftSegment := strings.Repeat(" ", 13)
	if left {
		leftSegment = strings.Repeat(" ", 7) + connectors.left + strings.Repeat(" ", 2)
	}
	middleSegment := strings.Repeat(" ", 2)
	if (connectors == upperConnectors && region.Neighbors[MapNorth].Label != "") || (connectors == lowerConnectors && region.Neighbors[MapSouth].Label != "") {
		middleSegment = "｜"
	}
	rightSegment := strings.Repeat(" ", 13)
	if right {
		rightSegment = strings.Repeat(" ", 2) + connectors.right + strings.Repeat(" ", 7)
	}
	return leftSegment + middleSegment + rightSegment
}
