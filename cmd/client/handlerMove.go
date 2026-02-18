package main

import (
	"fmt"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
)

// handlerMove accepts a game state struct and returns a new handler function
// that accepts a routing.ArmyMove struct.
func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(am gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(am)
	}
}
