package tui

import (
	"strings"

	"PMud/internal/client/layout"
	"PMud/internal/client/render"
	"PMud/internal/client/termwidth"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

const rightHUDWidth = 36

const defaultViewHeight = 26

const (
	roomPaneHeight    = 7
	minimapPaneHeight = 10
	statusPaneHeight  = 4
	inputPaneHeight   = 3
)

func View(model Model, catalog content.ClientCatalog, width int) layout.Block {
	return ViewWithSize(model, catalog, width, defaultViewHeight)
}

func ViewWithSize(model Model, catalog content.ClientCatalog, width int, height int) layout.Block {
	minHeight := inputPaneHeight + roomPaneHeight + minimapPaneHeight + statusPaneHeight
	if height < minHeight {
		height = minHeight
	}
	if width <= rightHUDWidth+8 {
		width = rightHUDWidth + 8
	}
	return renderMainFrame(model, catalog, width, height)
}

func renderMainFrame(model Model, catalog content.ClientCatalog, width int, height int) layout.Block {
	rightWidth := rightHUDWidth - 2
	leftWidth := width - rightWidth - 3
	if leftWidth < 1 {
		leftWidth = 1
	}
	roomSeparator := roomPaneHeight - 1
	minimapSeparator := minimapPaneHeight - 1
	statusSeparator := minimapSeparator + statusPaneHeight
	inputSeparator := height - inputPaneHeight
	leftContentWidth := leftWidth - 2
	rightContentWidth := rightWidth - 2
	leftRows := map[int]string{}
	rightRows := map[int]string{}
	putPane(leftRows, 1, leftContentWidth, paneBlock("房间 / 可见物", roomPaneBlock(model, catalog)).Lines)
	putPane(leftRows, roomSeparator+1, leftContentWidth, paneBlock("日志", eventHistoryBlock(model, catalog)).Lines)
	putPane(rightRows, 1, rightContentWidth, paneBlock("小地图", minimapPaneBlock(model, catalog)).Lines)
	putPane(rightRows, minimapSeparator+1, rightContentWidth, paneBlock("状态", statusPaneBlock()).Lines)
	putPane(rightRows, statusSeparator+1, rightContentWidth, paneBlock("当前任务", questPaneBlock(model)).Lines)

	out := make([]string, height)
	for y := range height {
		switch {
		case y == 0:
			out[y] = borderRow('┌', '┬', '┐', leftWidth, rightWidth)
		case y == height-1:
			out[y] = fullBorderRow('└', '┘', width)
		case y == inputSeparator:
			out[y] = inputSeparatorRow(width, leftWidth+1)
		case y > inputSeparator:
			out[y] = fullContentRow(promptLine(model, y-inputSeparator-1), width)
		case y == roomSeparator:
			out[y] = leftSeparatorRow(rightRows[y], leftWidth, rightContentWidth)
		case y == minimapSeparator || y == statusSeparator:
			out[y] = rightSeparatorRow(leftRows[y], leftContentWidth, rightWidth)
		default:
			out[y] = contentRow(leftRows[y], rightRows[y], leftContentWidth, rightContentWidth)
		}
	}
	return layout.NewBlock(out)
}

func borderRow(left rune, middle rune, right rune, leftWidth int, rightWidth int) string {
	return string(left) + strings.Repeat("─", leftWidth) + string(middle) + strings.Repeat("─", rightWidth) + string(right)
}

func fullBorderRow(left rune, right rune, width int) string {
	return string(left) + strings.Repeat("─", width-2) + string(right)
}

func inputSeparatorRow(width int, divider int) string {
	return "├" + strings.Repeat("─", divider-1) + "┴" + strings.Repeat("─", width-divider-2) + "┤"
}

func fullContentRow(content string, width int) string {
	return "│ " + termwidth.RightPad(content, width-4) + " │"
}

func promptLine(model Model, row int) string {
	if row == 0 {
		model = syncEditorFromInput(model)
		if model.ExitConfirmation {
			return "确认退出游戏？[y/N] " + model.Editor.RenderWithCursor()
		}
		return "> " + model.Editor.RenderWithCursor()
	}
	return ""
}

func leftSeparatorRow(right string, leftWidth int, rightContentWidth int) string {
	return "├" + strings.Repeat("─", leftWidth) + "┤ " + termwidth.RightPad(right, rightContentWidth) + " │"
}

func rightSeparatorRow(left string, leftContentWidth int, rightWidth int) string {
	return "│ " + termwidth.RightPad(left, leftContentWidth) + " ├" + strings.Repeat("─", rightWidth) + "┤"
}

func contentRow(left string, right string, leftContentWidth int, rightContentWidth int) string {
	return "│ " + termwidth.RightPad(left, leftContentWidth) + " │ " + termwidth.RightPad(right, rightContentWidth) + " │"
}

func putPane(rows map[int]string, startY int, width int, paneLines []string) {
	row := startY
	for _, line := range paneLines {
		wrapped := wrapVisible(line, width)
		for _, part := range wrapped {
			rows[row] = part
			row++
		}
	}
}

func wrapVisible(text string, width int) []string {
	if width < 1 {
		return []string{""}
	}
	if text == "" {
		return []string{""}
	}
	if termwidth.Width(text) <= width {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) > 1 {
		return wrapWords(words, width)
	}
	return hardWrapVisible(text, width)
}

func wrapWords(words []string, width int) []string {
	lines := make([]string, 0, 1)
	var builder strings.Builder
	for _, word := range words {
		wordWidth := termwidth.Width(word)
		if wordWidth > width {
			if builder.Len() > 0 {
				lines = append(lines, builder.String())
				builder.Reset()
			}
			wrapped := hardWrapVisible(word, width)
			lines = append(lines, wrapped[:len(wrapped)-1]...)
			builder.WriteString(wrapped[len(wrapped)-1])
			continue
		}
		separatorWidth := 0
		if builder.Len() > 0 {
			separatorWidth = 1
		}
		if termwidth.Width(builder.String())+separatorWidth+wordWidth > width {
			lines = append(lines, builder.String())
			builder.Reset()
		}
		if builder.Len() > 0 {
			builder.WriteByte(' ')
		}
		builder.WriteString(word)
	}
	lines = append(lines, builder.String())
	return lines
}

func hardWrapVisible(text string, width int) []string {
	lines := make([]string, 0, 1)
	var builder strings.Builder
	visibleWidth := 0
	for _, char := range text {
		charWidth := termwidth.Width(string(char))
		if visibleWidth+charWidth > width {
			lines = append(lines, builder.String())
			builder.Reset()
			visibleWidth = 0
		}
		builder.WriteRune(char)
		visibleWidth += charWidth
	}
	lines = append(lines, builder.String())
	return lines
}

func paneBlock(title string, body layout.Block) layout.Block {
	lines := make([]string, 0, len(body.Lines)+1)
	lines = append(lines, title)
	lines = append(lines, body.Lines...)
	return layout.NewBlock(lines)
}

func roomPaneBlock(model Model, catalog content.ClientCatalog) layout.Block {
	room := model.Regions.Room
	if room.Room == "" {
		return layout.NewBlock([]string{"尚未观察房间"})
	}
	return render.RenderBlock(protocol.Event{
		Name: "room",
		Fields: map[string]string{
			"room":            room.Room,
			"name_key":        room.NameKey,
			"description_key": room.DescriptionKey,
			"exits":           room.Exits,
			"items":           room.Items,
		},
	}, catalog)
}

func minimapPaneBlock(model Model, catalog content.ClientCatalog) layout.Block {
	region := MinimapRegion{
		AreaName: "当前区域",
		Current:  MinimapRoom{Label: minimapLabel(model, catalog)},
	}
	lines := []string{region.AreaName}
	lines = append(lines, renderMinimapGrid(region)...)
	return layout.NewBlock(lines)
}

func minimapLabel(model Model, catalog content.ClientCatalog) string {
	nameKey := model.Regions.Room.NameKey
	if nameKey == "" {
		return "当前"
	}
	name, ok := catalog.Text[content.TextKey(nameKey)]
	if !ok || name == "" {
		return "当前"
	}
	return name
}

func statusPaneBlock() layout.Block {
	return layout.NewBlock([]string{"状态系统未开放"})
}

func questPaneBlock(model Model) layout.Block {
	quest := model.Regions.Quest
	if quest.QuestID == "" {
		return layout.NewBlock([]string{"未追踪任务"})
	}
	return layout.NewBlock([]string{
		quest.QuestName,
		"阶段: " + quest.StageText,
		"目标: " + quest.Conditions,
	})
}

func eventHistoryBlock(model Model, catalog content.ClientCatalog) layout.Block {
	blocks := make([]layout.Block, 0, len(model.Events))
	for _, event := range model.Events {
		if event.Name == "room" || event.Name == "quest" {
			continue
		}
		blocks = append(blocks, render.RenderBlock(event, catalog))
	}
	return layout.VBox(blocks...)
}
