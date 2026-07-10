package presentation

type TextRenderer struct{}

func (t *TextRenderer) Render(event EchoEvent) string {
	return event.Line + "\n"
}
