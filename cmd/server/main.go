package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/cryptidcodes/learn-pub-sub/internal/gamelogic"
	"github.com/cryptidcodes/learn-pub-sub/internal/pubsub"
	"github.com/cryptidcodes/learn-pub-sub/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	var connectionString string = "amqp://guest:guest@localhost:5672/"

	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}

	// connection to RabbitMQ is now open
	defer connection.Close()
	fmt.Println("Connected to RabbitMQ")

	// create a channel on the connection
	ch, err := connection.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	defer ch.Close()

	gamelogic.PrintServerHelp()

	pubsub.DeclareAndBind(connection, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug, pubsub.Durable)

	// start an infinite REPL
	for i := 0; i == 0; {
		input := gamelogic.GetInput()
		if len(input) > 0 {
			command := input[0]
			switch command {
			case "pause":
				fmt.Println("Pausing game...")
				err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
				if err != nil {
					fmt.Println("Failed to publish pause message:", err)
					return
				}
			case "resume":
				fmt.Println("Resuming game...")
				err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
				if err != nil {
					fmt.Println("Failed to publish resume message:", err)
					return
				}
			case "quit":
				fmt.Println("Quitting game...")
				i = 1
			case "help":
				gamelogic.PrintServerHelp()
			default:
				fmt.Println("Unknown command. Type 'help' for a list of commands.")
			}
		}
	}

	// wait for ctrl-c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("\nShutting down Peril server...")
}
