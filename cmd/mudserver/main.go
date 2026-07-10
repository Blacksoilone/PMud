package main

import (
	"PMud/internal/session"
	"PMud/internal/world"
)

func main() {
	game := world.New()

	session.StartSession(game)

}
