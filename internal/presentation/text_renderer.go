package presentation

import "strings"

type TextRenderer struct{}

func (t *TextRenderer) RenderEcho(event EchoEvent) string {
	return event.Line + "\n"
}
func (t *TextRenderer) RenderInputDebug(event InputDebugEvent) string {
	return "Verb: " + event.Verb + "\nArgs: " + strings.Join(event.Args, ", ") + "\n"
}
