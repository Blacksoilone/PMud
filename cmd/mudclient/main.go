package main

import (
	"PMud/internal/client"
	"PMud/internal/content"
	"fmt"
	"net"
	"os"
)

const defaultAddress = "127.0.0.1:4000"

func main() {
	address := defaultAddress
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	conn, err := net.Dial("tcp", address)
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

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- client.RenderProtocolLines(conn, os.Stdout, compiled.Client)
	}()

	inputDone := make(chan error, 1)
	go func() {
		err := client.ForwardCommands(os.Stdin, conn)
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
