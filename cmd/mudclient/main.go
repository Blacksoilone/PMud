package main

import (
	"PMud/internal/client"
	"PMud/internal/content"
	"fmt"
	"io"
	"net"
	"os"
)

const defaultAddress = "127.0.0.1:4000"
const defaultTUIWidth = 80
const defaultTUIHistoryLimit = 20

type config struct {
	address string
	tui     bool
}

func main() {
	config := parseArgs(os.Args)

	conn, err := net.Dial("tcp", config.address)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	state := client.NewState(compiled.Client)

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- renderServer(conn, os.Stdout, state, config)
	}()

	inputDone := make(chan error, 1)
	go func() {
		err := client.ForwardResolvedCommands(os.Stdin, conn, state)
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			if closeErr := tcpConn.CloseWrite(); err == nil {
				err = closeErr
			}
		}
		inputDone <- err
	}()

	select {
	case err := <-serverDone:
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case err := <-inputDone:
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := <-serverDone; err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func parseArgs(args []string) config {
	config := config{address: defaultAddress}
	for _, arg := range args[1:] {
		if arg == "--tui" {
			config.tui = true
			continue
		}
		config.address = arg
	}
	return config
}

func renderServer(input io.Reader, output io.Writer, state *client.State, config config) error {
	if config.tui {
		return client.RenderTUIObservedProtocolLines(input, output, state, defaultTUIWidth, defaultTUIHistoryLimit)
	}
	return client.RenderObservedProtocolLines(input, output, state)
}
