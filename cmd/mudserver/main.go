package main

import (
	"PMud/internal/content"
	"PMud/internal/session"
	"PMud/internal/world"
)

func main() {
	game := buildGame()

	session.StartSession(game)

}

func buildGame() *world.World {
	compiled, err := content.Compile(content.TutorialSource())
	if err != nil {
		panic(err)
	}
	return world.NewFromSnapshot(compiled.Server, compiled.Client)
}
