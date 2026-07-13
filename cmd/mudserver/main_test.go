package main

import "testing"

func TestBuildGameLoadsTutorialData(t *testing.T) {
	game, err := buildGame("../../data/tutorial/source.json")
	if err != nil {
		t.Fatalf("buildGame: %v", err)
	}

	observation, ok := game.Look(game.StartRoom())
	if !ok {
		t.Fatal("expected start room to exist")
	}
	if observation.Name != "练习场入口" {
		t.Fatalf("start room name = %q, want 练习场入口", observation.Name)
	}
}
