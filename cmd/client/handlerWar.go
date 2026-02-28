package main

import (
	"fmt"
	"time"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/pubsub"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func publishGameLog(ch *amqp.Channel, rw gamelogic.RecognitionOfWar, message string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Username:    rw.Attacker.Username,
		Message:     message,
	}
	err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.GameLogSlug, rw.Attacker.Username), log)
	if err != nil {
		fmt.Println("Failed to publish game log message:", err)
		return err
	}
	return nil
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(rw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(rw)
		// debug print
		fmt.Println("War outcome:", warOutcome)
		fmt.Println("Winner:", winner)
		fmt.Println("Loser:", loser)
		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeYouWon:
			err := publishGameLog(ch, rw, fmt.Sprintf("%s won a war against %s", winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			err := publishGameLog(ch, rw, fmt.Sprintf("%s won a war against %s", winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := publishGameLog(ch, rw, fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser))
			if err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			fmt.Println("Error, unknown war outcome")
			return pubsub.NackDiscard
		}
	}
}
