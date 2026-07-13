package main

import (
	"PMud/internal/content"
	"PMud/internal/session"
	"PMud/internal/world"
	"fmt"
	"os"
)

const defaultContentPath = "data/tutorial/source.json"

func main() {
	game, err := buildGame(defaultContentPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	session.StartSession(game)

}

func buildGame(path string) (*world.World, error) {
	source, err := content.LoadSource(path)
	if err != nil {
		return nil, err
	}
	compiled, err := content.Compile(source)
	if err != nil {
		return nil, err
	}
	return world.NewFromSnapshot(compiled.Server, compiled.Client), nil
}
