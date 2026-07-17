package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"PMud/internal/client/keyinput"
	"PMud/internal/client/render"
	"PMud/internal/content"
	"PMud/internal/protocol"
)

var ErrProtocolLine = errors.New("protocol line")

const escapeSequenceTimeout = 40 * time.Millisecond

type tuiInputRead struct {
	data []byte
	err  error
}

func RenderProtocolLines(input io.Reader, output io.Writer, catalog content.ClientCatalog) error {
	return renderProtocolLines(input, output, catalog, nil)
}

func RenderObservedProtocolLines(input io.Reader, output io.Writer, state *State) error {
	return renderProtocolLines(input, output, state.catalog, state)
}

func RenderTUIProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error {
	return renderTUIProtocolLines(input, output, state, width, historyLimit, false)
}

func RenderTUIObservedProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int) error {
	return renderTUIProtocolLines(input, output, state, width, historyLimit, true)
}

func RenderTUIObservedProtocolLinesWithRuntime(input io.Reader, runtime *TUIRuntime) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if err := runtime.ObserveEvent(event); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func renderProtocolLines(input io.Reader, output io.Writer, catalog content.ClientCatalog, state *State) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if state != nil {
			state.Observe(event)
		}
		_, err = io.WriteString(output, render.Render(event, catalog))
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func renderTUIProtocolLines(input io.Reader, output io.Writer, state *State, width int, historyLimit int, observe bool) error {
	runtime := NewTUIRuntime(TUIRuntimeConfig{State: state, Output: output, Width: width, HistoryLimit: historyLimit})
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := protocol.ParseLine(scanner.Text())
		if err != nil {
			return fmt.Errorf("%w: %w", ErrProtocolLine, err)
		}
		if observe {
			err = runtime.ObserveEvent(event)
		} else {
			err = runtime.RenderEvent(event)
		}
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func ForwardCommands(input io.Reader, server io.Writer) error {
	return forwardCommands(input, server, nil)
}

func ForwardResolvedCommands(input io.Reader, server io.Writer, state *State) error {
	return forwardCommands(input, server, state)
}

func ForwardTUILines(input io.Reader, server io.Writer, runtime *TUIRuntime) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if err := runtime.SubmitLine(scanner.Text(), server); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func ForwardTUIKeyInput(input io.Reader, server io.Writer, runtime *TUIRuntime) error {
	var decoder keyinput.Decoder
	reads := readTUIInput(input)
	timer := time.NewTimer(escapeSequenceTimeout)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()
	var escapeTimeout <-chan time.Time
	for {
		select {
		case read := <-reads:
			if len(read.data) > 0 {
				if err := forwardTUIKeyActions(decoder.Feed(read.data), server, runtime); err != nil {
					return err
				}
			}
			if decoder.HasStandaloneEscape() {
				resetTimer(timer, escapeSequenceTimeout)
				escapeTimeout = timer.C
			} else {
				stopTimer(timer)
				escapeTimeout = nil
			}
			if read.err == io.EOF {
				if err := forwardTUIKeyActions(decoder.Flush(), server, runtime); err != nil {
					return err
				}
				return nil
			}
			if read.err != nil {
				return read.err
			}
		case <-escapeTimeout:
			escapeTimeout = nil
			if err := forwardTUIKeyActions(decoder.Flush(), server, runtime); err != nil {
				return err
			}
		}
	}
}

func readTUIInput(input io.Reader) <-chan tuiInputRead {
	reads := make(chan tuiInputRead, 1)
	go func() {
		defer close(reads)
		buffer := make([]byte, 256)
		for {
			count, err := input.Read(buffer)
			data := append([]byte(nil), buffer[:count]...)
			reads <- tuiInputRead{data: data, err: err}
			if err != nil {
				return
			}
		}
	}()
	return reads
}

func resetTimer(timer *time.Timer, timeout time.Duration) {
	stopTimer(timer)
	timer.Reset(timeout)
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func forwardTUIKeyActions(actions []keyinput.Action, server io.Writer, runtime *TUIRuntime) error {
	for _, action := range actions {
		if err := runtime.ApplyInput(action.Input, server); err != nil {
			return err
		}
	}
	return nil
}

func forwardCommands(input io.Reader, server io.Writer, state *State) error {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()
		if state != nil {
			resolution := state.ResolveCommandInput(line)
			if !resolution.Send {
				continue
			}
			line = resolution.Command
		}
		_, err := io.WriteString(server, line+"\n")
		if err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
