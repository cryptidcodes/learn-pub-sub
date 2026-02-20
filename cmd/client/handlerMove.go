package main

import (
	"fmt"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/pubsub"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

// handlerMove accepts a game state struct and returns a new handler function
// that accepts a routing.ArmyMove struct.
func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {
	return func(move gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		moveOutcome := gs.HandleMove(move)
		switch moveOutcome {
		case gamelogic.MoveOutcomeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			// publish a message to the topic exchange with the routing key $WARPREFIX.$USERNAME
			err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, gs.GetPlayerSnap().Username), gamelogic.RecognitionOfWar{
				Attacker: move.Player,
				Defender: gs.GetPlayerSnap(),
			})
			if err != nil {
				fmt.Println("Failed to publish war message:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}
