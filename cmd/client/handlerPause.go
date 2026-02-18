package main

import (
	"fmt"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
)

// handlerPause accepts a game state struct and returns a new handler function
// that accepts a routing.PlayingState struct.
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}
