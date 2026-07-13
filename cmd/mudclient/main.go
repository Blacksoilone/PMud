package main

import (
	"PMud/internal/client"
	"PMud/internal/client/rawterm"
	"PMud/internal/content"
	"fmt"
	"net"
	"os"
)

const defaultAddress = "127.0.0.1:4000"
const defaultContentPath = "data/tutorial/source.json"
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

	catalog, err := loadClientCatalog(defaultContentPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	state := client.NewState(catalog)
	var runtime *client.TUIRuntime
	if config.tui {
		runtime = client.NewTUIRuntime(client.TUIRuntimeConfig{
			State:        state,
			Output:       os.Stdout,
			Width:        defaultTUIWidth,
			HistoryLimit: defaultTUIHistoryLimit,
		})
		session, err := rawterm.Start(int(os.Stdin.Fd()), rawterm.RealController())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer session.Close()
	}

	serverDone := make(chan error, 1)
	go func() {
		if config.tui {
			serverDone <- client.RenderTUIObservedProtocolLinesWithRuntime(conn, runtime)
			return
		}
		serverDone <- client.RenderObservedProtocolLines(conn, os.Stdout, state)
	}()

	inputDone := make(chan error, 1)
	go func() {
		var err error
		if config.tui {
			err = client.ForwardTUIKeyInput(os.Stdin, conn, runtime)
		} else {
			err = client.ForwardResolvedCommands(os.Stdin, conn, state)
		}
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

func loadClientCatalog(path string) (content.ClientCatalog, error) {
	source, err := content.LoadSource(path)
	if err != nil {
		return content.ClientCatalog{}, err
	}
	compiled, err := content.Compile(source)
	if err != nil {
		return content.ClientCatalog{}, err
	}
	return compiled.Client, nil
}
