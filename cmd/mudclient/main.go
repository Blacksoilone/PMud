package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"PMud/internal/client"
	"PMud/internal/client/rawterm"
	"PMud/internal/client/screen"
	"PMud/internal/content"
)

const (
	defaultAddress         = "127.0.0.1:4000"
	defaultContentPath     = "data/tutorial/source.json"
	defaultTUIWidth        = 80
	defaultTUIHeight       = 26
	defaultTUIHistoryLimit = 20
)

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
	var resizeDone chan struct{}
	if config.tui {
		controller := rawterm.RealController()
		session, err := rawterm.Start(int(os.Stdin.Fd()), controller)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer session.Close()

		width, height, err := controller.Size(int(os.Stdout.Fd()))
		if err != nil {
			width = defaultTUIWidth
			height = defaultTUIHeight
		}
		if err := screen.EnterAlternateScreen(os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer func() {
			_ = screen.ExitAlternateScreen(os.Stdout)
		}()

		runtime = client.NewTUIRuntime(client.TUIRuntimeConfig{
			State:        state,
			Output:       os.Stdout,
			Width:        width,
			Height:       height,
			HistoryLimit: defaultTUIHistoryLimit,
		})
		resizeSignals := make(chan os.Signal, 1)
		resizeDone = make(chan struct{})
		signal.Notify(resizeSignals, syscall.SIGWINCH)
		defer func() {
			signal.Stop(resizeSignals)
			close(resizeDone)
		}()
		go func() {
			for {
				select {
				case <-resizeSignals:
					width, height, sizeErr := controller.Size(int(os.Stdout.Fd()))
					if sizeErr == nil {
						_ = runtime.Resize(width, height)
					}
				case <-resizeDone:
					return
				}
			}
		}()
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
		if errors.Is(err, client.ErrTUIExit) {
			return
		}
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
	config := config{address: defaultAddress, tui: true}
	for _, arg := range args[1:] {
		if arg == "--tui" {
			config.tui = true
			continue
		}
		if arg == "--line" {
			config.tui = false
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
