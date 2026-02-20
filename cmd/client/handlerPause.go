package main

import (
	"fmt"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/pubsub"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
)

// handlerPause accepts a game state struct and returns a new handler function
// that accepts a routing.PlayingState struct.
func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {
	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}
